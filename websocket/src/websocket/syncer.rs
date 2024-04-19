use futures::future::join_all;
use std::{collections::HashMap, ops::ControlFlow};
use tokio::sync::mpsc::{Receiver, Sender};

use sqlx::PgPool;
use tracing::{instrument, Instrument};
use uuid::Uuid;

use crate::document::Document;

use super::Message;

pub struct Syncer {
    clients: HashMap<Uuid, Sender<Vec<u8>>>,
    document: Document,
    rx: Receiver<Message>,
    pool: PgPool,
}

impl Syncer {
    pub fn new(pool: PgPool, document: Document, rx: Receiver<Message>) -> Self {
        Self {
            clients: HashMap::new(),
            rx,
            pool,
            document,
        }
    }

    #[instrument(name="Syncer", skip(self), fields(document_id=%self.document.id))]
    pub fn run(mut self) {
        tokio::spawn(
            async move {
                tracing::info!("starting syncer");
                while let Some(message) = self.rx.recv().await {
                    if self.process_message(message).await.is_break() {
                        break;
                    };
                }
            }
            .instrument(tracing::Span::current()),
        );
    }

    async fn process_message(&mut self, message: Message) -> ControlFlow<(), ()> {
        match message {
            Message::Connect(id, tx) => {
                self.clients.insert(id, tx);
            }
            Message::Disconnect(id) => {
                self.clients.remove(&id);
                if self.clients.is_empty() {
                    return ControlFlow::Break(());
                };
            }
            Message::Sync(id, update) => {
                join_all(
                    self.clients
                        .iter()
                        .filter(|(client_id, _)| **client_id != id)
                        .map(|(_, tx)| tx.send(update.clone()))
                        .collect::<Vec<_>>(),
                )
                .await;
            }
        };
        ControlFlow::Continue(())
    }
}

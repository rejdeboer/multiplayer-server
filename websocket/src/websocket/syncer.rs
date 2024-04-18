use std::collections::HashMap;
use tokio::sync::mpsc::{Receiver, Sender};

use sqlx::PgPool;
use tracing::{instrument, Instrument};
use uuid::Uuid;

use crate::document::Document;

use super::Message;

pub struct Syncer {
    clients: HashMap<Uuid, Sender<Message>>,
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
                    self.process_message(message).await;
                }
            }
            .instrument(tracing::Span::current()),
        );
    }

    async fn process_message(&self, message: Message) {}
}

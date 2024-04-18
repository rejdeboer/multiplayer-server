use std::collections::HashMap;
use tokio::sync::mpsc::{Receiver, Sender};

use sqlx::PgPool;
use uuid::Uuid;

use crate::document::Document;

pub struct Syncer {
    clients: HashMap<Uuid, Sender<String>>,
    document: Document,
    rx: Receiver<String>,
    pool: PgPool,
}

impl Syncer {
    pub fn new(pool: PgPool, document: Document, rx: Receiver<String>) -> Self {
        Self {
            clients: HashMap::new(),
            rx,
            pool,
            document,
        }
    }

    pub fn run(mut self) {
        tokio::spawn(async move {
            while let Some(message) = self.rx.recv().await {
                self.process_message(message).await;
            }
        });
    }

    async fn process_message(&self, message: String) {}
}

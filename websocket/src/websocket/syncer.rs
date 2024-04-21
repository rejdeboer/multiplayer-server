use futures::future::join_all;
use std::{collections::HashMap, ops::ControlFlow};
use tokio::sync::mpsc::{Receiver, Sender};
use yrs::{updates::decoder::Decode, Doc, ReadTxn, StateVector, Transact, Update};

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
            Message::Sync(id, mut update) => {
                join_all(
                    self.clients
                        .iter()
                        .filter(|(client_id, _)| **client_id != id)
                        .map(|(_, tx)| tx.send(update.clone()))
                        .collect::<Vec<_>>(),
                )
                .await;

                // Remove message type
                update.pop();

                self.store_update(update).await;
            }
            Message::GetDiff(id, mut state_vector) => {
                tracing::info!(%id, "received GetDiff message");
                let client_tx = self.clients.get(&id).expect("get client_tx").clone();

                // Pop the message type
                state_vector.pop();
                let state_vector =
                    StateVector::decode_v1(&state_vector).expect("state vector decoded");

                compute_and_send_diff(client_tx, state_vector, self.document.id, self.pool.clone())
                    .await;
            }
        };
        ControlFlow::Continue(())
    }

    async fn store_update(&self, update: Vec<u8>) {
        let pool = self.pool.clone();
        let document_id = self.document.id;

        let mut txn = pool
            .begin()
            .await
            .expect("receive pool connection for update transaction");

        let current_clock = sqlx::query!(
            r#"
                SELECT COALESCE(MAX(clock), -1) as value 
                FROM document_updates
                WHERE document_id = $1;
                "#,
            document_id
        )
        .fetch_one(&mut *txn)
        .await
        .expect("retrieve current clock")
        .value
        .unwrap();

        sqlx::query!(
            r#"
                INSERT INTO document_updates (document_id, clock, value)
                VALUES($1, $2, $3);
                "#,
            document_id,
            current_clock + 1,
            update
        )
        .execute(&mut *txn)
        .await
        .expect("update stored");

        txn.commit().await.expect("transcation committed");
    }
}

async fn compute_and_send_diff(
    client_tx: Sender<Vec<u8>>,
    state_vector: StateVector,
    document_id: Uuid,
    pool: PgPool,
) {
    let encoded_updates = get_document_updates(document_id, pool).await;

    let mut update = compute_diff(state_vector, encoded_updates);
    update.push(super::MESSAGE_SYNC);

    client_tx
        .send(update)
        .await
        .expect("document updates sent to client");
}

async fn get_document_updates(document_id: Uuid, pool: PgPool) -> Vec<Vec<u8>> {
    sqlx::query!(
        r#"
        SELECT value
        FROM document_updates
        WHERE document_id = $1
        ORDER BY clock;
        "#,
        document_id
    )
    .fetch_all(&pool)
    .await
    .expect("retrieve document updates")
    .into_iter()
    .map(|update| update.value)
    .collect::<_>()
}

fn compute_diff(state_vector: StateVector, encoded_updates: Vec<Vec<u8>>) -> Vec<u8> {
    let updates = encoded_updates
        .into_iter()
        .map(|update| Update::decode_v1(&update).expect("update decoded"))
        .collect::<Vec<Update>>();

    let merged = Update::merge_updates(updates);
    let doc = Doc::new();
    let mut txn = doc.transact_mut();
    txn.apply_update(merged);

    txn.encode_diff_v1(&state_vector)
}

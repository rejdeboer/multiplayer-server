mod client;
mod syncer;

use axum::extract::ws::WebSocket;
use tokio::sync::mpsc::{channel, Sender};
use uuid::Uuid;

use crate::auth::User;

use self::client::Client;
pub use syncer::Syncer;

pub async fn handle_socket(socket: WebSocket, user: User, doc_handle: Sender<Message>) {
    let (client_tx, client_rx) = channel(128);
    let client = Client::new(user, client_tx, doc_handle);

    client.run(socket, client_rx).await;
}

#[derive(Clone)]
pub enum Message {
    Connect(Uuid, Sender<Vec<u8>>),
    Disconnect(Uuid),
    Sync(Uuid, Vec<u8>),
}

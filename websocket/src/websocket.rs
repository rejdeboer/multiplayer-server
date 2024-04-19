mod client;
mod syncer;

use axum::extract::ws::WebSocket;
use tokio::sync::mpsc::{channel, Sender};
use uuid::Uuid;

use crate::auth::User;

use self::client::Client;
pub use syncer::Syncer;

// WebSocket message type flags
pub const MESSAGE_SYNC: u8 = 0;
pub const MESSAGE_GET_DIFF: u8 = 1;

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
    GetDiff(Uuid, Vec<u8>),
}

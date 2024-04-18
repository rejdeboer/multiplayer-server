mod client;
mod syncer;

use axum::extract::ws::WebSocket;
use tokio::sync::mpsc::Sender;
use uuid::Uuid;

use crate::auth::User;

use self::client::Client;
pub use syncer::Syncer;

pub async fn handle_socket(socket: WebSocket, user: User, doc_handle: Sender<Message>) {
    let mut client = Client::new(socket, user, doc_handle);

    client.run().await;
}

pub enum Message {
    Connect(Uuid),
    Disconnect(Uuid),
}

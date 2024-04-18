mod client;
pub mod syncer;

use axum::extract::ws::WebSocket;
use tokio::sync::mpsc::Sender;

use crate::auth::User;

use self::client::Client;

pub async fn handle_socket(socket: WebSocket, user: User, doc_handle: Sender<String>) {
    let mut client = Client::new(socket, user, doc_handle);

    client.run().await;
}

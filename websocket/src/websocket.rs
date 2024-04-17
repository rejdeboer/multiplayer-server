mod client;
pub mod room;

use std::sync::Arc;

use axum::extract::ws::WebSocket;

use crate::{auth::User, document::Document, startup::ApplicationState};

use self::{client::Client, room::Room};

pub async fn handle_socket(
    socket: WebSocket,
    user: User,
    document: Document,
    state: Arc<ApplicationState>,
) {
    let mut room = state
        .rooms
        .lock()
        .expect("received rooms lock")
        .entry(document.id)
        .or_insert(Room::new(state.pool.clone()));

    let mut client = Client::new(socket, user);
    client.run().await;
}

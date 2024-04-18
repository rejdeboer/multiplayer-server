use std::ops::ControlFlow;

use axum::extract::ws::{Message, WebSocket};
use tokio::sync::mpsc::{channel, Receiver, Sender};
use tracing::instrument;
use uuid::Uuid;

use crate::auth::User;

#[derive(Debug)]
pub struct Client {
    tx: Sender<String>,
    id: Uuid,
    rx: Receiver<String>,
    doc_handle: Sender<String>,
    socket: WebSocket,
    user: User,
}

impl Client {
    pub fn new(socket: WebSocket, user: User, doc_handle: Sender<String>) -> Self {
        let (tx, rx) = channel(128);
        Self {
            id: Uuid::new_v4(),
            tx,
            rx,
            doc_handle,
            socket,
            user,
        }
    }

    #[instrument(name="websocket connection", skip(self), fields(user = ?self.user))]
    pub async fn run(&mut self) {
        tracing::info!("new client connected");
        while let Some(Ok(msg)) = self.socket.recv().await {
            if self.process_message(msg).is_break() {
                break;
            }
        }
    }

    fn process_message(&self, msg: Message) -> ControlFlow<(), ()> {
        match msg {
            Message::Binary(d) => {
                tracing::debug!(content=?d, "received bytes");
            }
            Message::Close(c) => {
                if let Some(cf) = c {
                    tracing::info!(code = %cf.code, reason = %cf.reason, "received close message");
                } else {
                    tracing::warn!("somehow received close message without CloseFrame");
                }
                return ControlFlow::Break(());
            }
            Message::Pong(_) => (),
            Message::Ping(_) => (),
            msg => tracing::warn!(?msg, "unhandled message"),
        }
        ControlFlow::Continue(())
    }
}

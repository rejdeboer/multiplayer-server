use std::ops::ControlFlow;

use axum::extract::ws::{Message as WSMessage, WebSocket};
use tokio::sync::mpsc::{channel, Receiver, Sender};
use tracing::instrument;
use uuid::Uuid;

use crate::auth::User;

use super::Message;

#[derive(Debug)]
pub struct Client {
    tx: Sender<Message>,
    id: Uuid,
    rx: Receiver<Message>,
    doc_handle: Sender<Message>,
    socket: WebSocket,
    user: User,
}

impl Client {
    pub fn new(socket: WebSocket, user: User, doc_handle: Sender<Message>) -> Self {
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

    #[instrument(name="websocket client", skip(self), fields(user = ?self.user))]
    pub async fn run(&mut self) {
        self.doc_handle
            .send(Message::Connect(self.id))
            .await
            .expect("client connects to syncer");
        tracing::info!("new client connected");

        while let Some(Ok(msg)) = self.socket.recv().await {
            if self.process_message(msg).is_break() {
                break;
            }
        }

        self.doc_handle
            .send(Message::Disconnect(self.id))
            .await
            .expect("client disconnects from syncer");
        tracing::info!("client disconnected");
    }

    fn process_message(&self, msg: WSMessage) -> ControlFlow<(), ()> {
        match msg {
            WSMessage::Binary(d) => {
                tracing::debug!(content=?d, "received bytes");
            }
            WSMessage::Close(c) => {
                if let Some(cf) = c {
                    tracing::info!(code = %cf.code, reason = %cf.reason, "received close message");
                } else {
                    tracing::warn!("somehow received close message without CloseFrame");
                }
                return ControlFlow::Break(());
            }
            WSMessage::Pong(_) => (),
            WSMessage::Ping(_) => (),
            msg => tracing::warn!(?msg, "unhandled message"),
        }
        ControlFlow::Continue(())
    }
}

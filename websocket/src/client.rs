use std::ops::ControlFlow;

use axum::extract::ws::{Message, WebSocket};
use sqlx::PgPool;

#[derive(Debug)]
pub struct Client {
    socket: WebSocket,
    pool: PgPool,
}

impl Client {
    pub fn new(socket: WebSocket, pool: PgPool) -> Self {
        Self { socket, pool }
    }

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

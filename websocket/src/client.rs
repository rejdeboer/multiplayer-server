use std::{net::SocketAddr, ops::ControlFlow};

use axum::extract::ws::{Message, WebSocket};
use sqlx::PgPool;

pub struct Client {
    socket: WebSocket,
    who: SocketAddr,
    pool: PgPool,
}

impl Client {
    pub fn new(socket: WebSocket, who: SocketAddr, pool: PgPool) -> Self {
        Self { socket, who, pool }
    }

    pub async fn run(&mut self) {
        while let Some(Ok(msg)) = self.socket.recv().await {
            if self.process_message(msg).is_break() {
                break;
            }
        }
    }

    fn process_message(&self, msg: Message) -> ControlFlow<(), ()> {
        let who = &self.who;
        match msg {
            Message::Binary(d) => {
                println!(">>> {} sent {} bytes: {:?}", who, d.len(), d);
            }
            Message::Close(c) => {
                if let Some(cf) = c {
                    println!(
                        ">>> {} sent close with code {} and reason `{}`",
                        who, cf.code, cf.reason
                    );
                } else {
                    println!(">>> {who} somehow sent close message without CloseFrame");
                }
                return ControlFlow::Break(());
            }
            Message::Pong(_) => (),
            Message::Ping(_) => (),
            msg => tracing::warn!("unhandled message: {:?}", msg),
        }
        ControlFlow::Continue(())
    }
}

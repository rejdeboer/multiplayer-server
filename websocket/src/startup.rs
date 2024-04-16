use axum::{
    extract::{ws::WebSocket, ConnectInfo, State, WebSocketUpgrade},
    response::IntoResponse,
    routing::get,
    Router,
};
use axum_extra::TypedHeader;
use sqlx::{postgres::PgPoolOptions, PgPool};
use std::net::SocketAddr;
use tokio::net::TcpListener;
use tower_http::trace::{DefaultMakeSpan, TraceLayer};

use crate::{
    client::Client,
    configuration::{DatabaseSettings, Settings},
};

pub struct Application {
    listener: TcpListener,
    router: Router,
}

impl Application {
    pub async fn build(settings: Settings) -> Result<Self, std::io::Error> {
        let address = format!(
            "{}:{}",
            settings.application.host, settings.application.port
        );

        let listener = TcpListener::bind(address).await.unwrap();
        let connection_pool = get_connection_pool(&settings.database);

        let router = Router::new()
            .route("/ws", get(ws_handler))
            .layer(
                TraceLayer::new_for_http()
                    .make_span_with(DefaultMakeSpan::default().include_headers(true)),
            )
            .with_state(connection_pool);

        Ok(Self { listener, router })
    }

    pub async fn run_until_stopped(self) -> Result<(), std::io::Error> {
        tracing::info!("listening on {}", self.listener.local_addr().unwrap());
        axum::serve(
            self.listener,
            self.router
                .into_make_service_with_connect_info::<SocketAddr>(),
        )
        .await
    }
}

pub fn get_connection_pool(settings: &DatabaseSettings) -> PgPool {
    PgPoolOptions::new().connect_lazy_with(settings.with_db())
}

// TODO: Add connection context for tracing
async fn ws_handler(
    ws: WebSocketUpgrade,
    user_agent: Option<TypedHeader<headers::UserAgent>>,
    ConnectInfo(_addr): ConnectInfo<SocketAddr>,
    State(pool): State<PgPool>,
) -> impl IntoResponse {
    let _user_agent = if let Some(TypedHeader(user_agent)) = user_agent {
        user_agent.to_string()
    } else {
        String::from("Unknown client")
    };
    ws.on_upgrade(move |socket| handle_socket(socket, pool))
}

async fn handle_socket(socket: WebSocket, pool: PgPool) {
    let mut client = Client::new(socket, pool);
    client.run().await;
}

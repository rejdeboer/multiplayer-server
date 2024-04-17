use axum::{
    extract::{ws::WebSocket, ConnectInfo, Path, State, WebSocketUpgrade},
    middleware,
    response::Response,
    routing::get,
    Extension, Router,
};
use axum_extra::TypedHeader;
use sqlx::{postgres::PgPoolOptions, PgPool};
use std::{net::SocketAddr, str::FromStr};
use tokio::net::TcpListener;
use tower_http::trace::{DefaultMakeSpan, TraceLayer};
use uuid::Uuid;

use crate::{
    auth::{auth_middleware, User},
    client::Client,
    configuration::{DatabaseSettings, Settings},
    document::Document,
    error::ApiError,
};

pub struct Application {
    listener: TcpListener,
    router: Router,
    port: u16,
}

#[derive(Clone)]
pub struct ApplicationState {
    pool: PgPool,
}

impl Application {
    pub async fn build(settings: Settings) -> Result<Self, std::io::Error> {
        let address = format!(
            "{}:{}",
            settings.application.host, settings.application.port
        );

        let listener = TcpListener::bind(address).await.unwrap();
        let port = listener.local_addr().unwrap().port();
        let connection_pool = get_connection_pool(&settings.database);

        let application_state = ApplicationState {
            pool: connection_pool,
        };

        let router = Router::new()
            .route("/sync/:document_id", get(ws_handler))
            .route_layer(middleware::from_fn_with_state(
                settings.application.signing_key,
                auth_middleware,
            ))
            .layer(
                TraceLayer::new_for_http()
                    .make_span_with(DefaultMakeSpan::default().include_headers(true)),
            )
            .with_state(application_state);

        Ok(Self {
            listener,
            router,
            port,
        })
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

    pub fn port(&self) -> u16 {
        self.port
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
    State(state): State<ApplicationState>,
    Path(document_id): Path<String>,
    Extension(user): Extension<User>,
) -> Result<Response, ApiError> {
    let _user_agent = if let Some(TypedHeader(user_agent)) = user_agent {
        user_agent.to_string()
    } else {
        String::from("Unknown client")
    };

    let document_id = Uuid::from_str(&document_id)
        .map_err(|_| ApiError::BadRequest("please provide a valid document UUID".to_string()))?;

    let document = sqlx::query_as!(
        Document,
        r#"
        SELECT id, owner_id, name, state_vector
        FROM documents
        WHERE id = $1 
        "#,
        document_id
    )
    .fetch_one(&state.pool)
    .await
    .map_err(|_| ApiError::DocumentNotFoundError(document_id))?;

    if document.owner_id != user.id {
        tracing::error!(
            ?user,
            document = %document_id,
            "user does not have access to document"
        );
        return Err(ApiError::DocumentNotFoundError(document_id));
    }

    Ok(ws.on_upgrade(move |socket| handle_socket(socket, user, document, state)))
}

async fn handle_socket(socket: WebSocket, user: User, document: Document, state: ApplicationState) {
    let mut client = Client::new(socket, user, state.pool);
    client.run().await;
}

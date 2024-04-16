use once_cell::sync::Lazy;
use sqlx::types::Uuid;
use sqlx::{Connection, Executor, PgConnection, PgPool};
use websocket::configuration::{get_configuration, DatabaseSettings};
use websocket::startup::{get_connection_pool, Application};
use websocket::telemetry::{get_subscriber, init_subscriber};

static TRACING: Lazy<()> = Lazy::new(|| {
    if std::env::var("TEST_LOG").is_ok() {
        let subscriber = get_subscriber();
        init_subscriber(subscriber);
    } else {
        let subscriber = get_subscriber();
        init_subscriber(subscriber);
    };
});

pub struct TestApp {
    pub address: String,
    pub port: u16,
    pub db_pool: PgPool,
}

impl TestApp {
    pub async fn test_document_id(&self) -> String {
        let row = sqlx::query!("SELECT id FROM documents LIMIT 1")
            .fetch_one(&self.db_pool)
            .await
            .expect("fetched document");
        row.id.to_string()
    }
}

pub async fn spawn_app() -> TestApp {
    // Only initialize tracer once instead of every test
    Lazy::force(&TRACING);

    let configuration = {
        let mut c = get_configuration().expect("configuration fetched");
        c.database.db_name = Uuid::new_v4().to_string();
        c.application.port = 0;
        c
    };

    configure_database(&configuration.database).await;
    let application = Application::build(configuration.clone())
        .await
        .expect("application built");
    let application_port = application.port();
    let _ = tokio::spawn(application.run_until_stopped());

    let test_app = TestApp {
        address: format!("http://localhost:{}", application_port),
        port: application_port,
        db_pool: get_connection_pool(&configuration.database),
    };

    add_test_document(&test_app.db_pool).await;
    test_app
}

async fn configure_database(config: &DatabaseSettings) -> PgPool {
    let mut connection = PgConnection::connect_with(&config.without_db())
        .await
        .expect("connected to postgres");
    connection
        .execute(format!(r#"CREATE DATABASE "{}";"#, config.db_name).as_str())
        .await
        .expect("db created");

    let connection_pool = PgPool::connect_with(config.with_db())
        .await
        .expect("Failed to connect to Postgres.");
    sqlx::migrate!("../db/migrations")
        .run(&connection_pool)
        .await
        .expect("migration successful");

    connection_pool
}

async fn add_test_document(pool: &PgPool) {
    sqlx::query!(
        "INSERT INTO documents (id, name, owner_id, state_vector)
        VALUES ($1, $2, $3, $4)",
        Uuid::new_v4(),
        Uuid::new_v4().to_string(),
        Uuid::new_v4(),
        vec![],
    )
    .execute(pool)
    .await
    .expect("test document created");
}

use sqlx::PgPool;

pub struct Room {
    pool: PgPool,
}

impl Room {
    pub fn new(pool: PgPool) -> Self {
        Self { pool }
    }
}

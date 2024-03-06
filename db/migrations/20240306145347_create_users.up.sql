CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT UUID_GENERATE_V4(),
	email text NOT NULL,
	passhash text NOT NULL
);

CREATE INDEX IF NOT EXISTS "idx_users_id" ON "users" ("id");

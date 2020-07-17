CREATE TABLE IF NOT EXISTS users (
    uuid uuid DEFAULT uuid_generate_v4 (),
    email VARCHAR(255) NOT NULL CHECK (email <> '') UNIQUE,
    password VARCHAR(255) NOT NULL CHECK (password <> ''),
    new_password VARCHAR(255) CHECK (password <> ''),
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    salt VARCHAR(255) NOT NULL default md5(random()::text || clock_timestamp()::text),
    created_at TIMESTAMPTZ NOT NULL default current_timestamp,
    updated_at TIMESTAMPTZ NOT NULL default current_timestamp,
    PRIMARY KEY (uuid)
);

CREATE TRIGGER set_timestamp BEFORE UPDATE ON users FOR EACH ROW EXECUTE PROCEDURE  trigger_set_timestamp();
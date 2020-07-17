
CREATE TABLE IF NOT EXISTS profiles (
    uuid uuid DEFAULT uuid_generate_v4 (),
    user_uuid uuid NOT NULL REFERENCES users(uuid) ON DELETE RESTRICT UNIQUE,
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    address VARCHAR(255),
    phone VARCHAR(21),
    gender CHAR(1) CHECK (gender IN ('m', 'f')),
    dob DATE,
    created_at TIMESTAMPTZ NOT NULL default current_timestamp,
    updated_at TIMESTAMPTZ NOT NULL default current_timestamp,
    PRIMARY KEY (uuid)
);

CREATE TRIGGER set_timestamp BEFORE UPDATE ON profiles FOR EACH ROW EXECUTE PROCEDURE  trigger_set_timestamp();
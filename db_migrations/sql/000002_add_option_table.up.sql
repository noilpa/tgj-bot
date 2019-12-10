CREATE TABLE IF NOT EXISTS options (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE,
    item JSON NOT NULL,
    updated_at timestamp without time zone
);

INSERT INTO options (name, item, updated_at) VALUES('last_send_notify', '{"value": 0}', now());

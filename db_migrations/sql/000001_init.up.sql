CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    telegram_id TEXT UNIQUE,
    telegram_username TEXT UNIQUE,
    gitlab_id TEXT UNIQUE,
    gitlab_name TEXT NOT NULL DEFAULT '',
    jira_id TEXT,
    is_active BOOLEAN,
    role TEXT
);

CREATE TABLE IF NOT EXISTS mrs (
    id SERIAL PRIMARY KEY,
    url TEXT UNIQUE,
    author_id INTEGER NOT NULL,
    is_closed BOOLEAN DEFAULT FALSE,
    jira_id INTEGER NOT NULL DEFAULT 0,
    jira_priority INTEGER NOT NULL DEFAULT 0,
    gitlab_id INTEGER NOT NULL DEFAULT 0,
    jira_status INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY(author_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS reviews (
    mr_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    is_approved BOOLEAN DEFAULT FALSE,
    is_commented BOOLEAN DEFAULT FALSE,
    updated_at BIGINT,
    PRIMARY KEY (mr_id, user_id),
    FOREIGN KEY(mr_id) REFERENCES mrs(id),
    FOREIGN KEY(user_id) REFERENCES users(id)
);

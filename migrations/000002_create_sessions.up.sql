CREATE TABLE IF NOT EXISTS sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    work_name TEXT,
    start_time TIMESTAMP NOT NULL,
    planned_end TIMESTAMP NOT NULL,
    actual_end TIMESTAMP,
    icon_id INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_start_time ON sessions(start_time DESC);
CREATE INDEX idx_sessions_active ON sessions(user_id, actual_end) WHERE actual_end IS NULL;

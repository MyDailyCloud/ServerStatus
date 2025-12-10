CREATE TABLE IF NOT EXISTS server_info (
    id SERIAL PRIMARY KEY,
    session_id TEXT UNIQUE NOT NULL,
    hostname TEXT NOT NULL,
    project_key TEXT NOT NULL,
    latest_data TEXT,
    last_seen TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS history_data (
    id SERIAL PRIMARY KEY,
    session_id TEXT NOT NULL,
    hostname TEXT NOT NULL,
    project_key TEXT NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    data TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_history_session_ts ON history_data(session_id, timestamp);
CREATE INDEX IF NOT EXISTS idx_history_hostname_ts ON history_data(hostname, timestamp);
CREATE INDEX IF NOT EXISTS idx_history_project ON history_data(project_key);

CREATE TABLE IF NOT EXISTS access_key_cache (
    id SERIAL PRIMARY KEY,
    cache_key TEXT UNIQUE NOT NULL,
    access_key TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_access_key_cache_key ON access_key_cache(cache_key);

CREATE TABLE IF NOT EXISTS visitor_events (
    id SERIAL PRIMARY KEY,
    project_key TEXT NOT NULL,
    domain TEXT,
    page_url TEXT,
    referrer TEXT,
    user_agent TEXT,
    ip TEXT,
    session_id TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_visitor_project_time ON visitor_events(project_key, created_at);
CREATE INDEX IF NOT EXISTS idx_visitor_page ON visitor_events(page_url);
CREATE INDEX IF NOT EXISTS idx_visitor_domain ON visitor_events(domain);

CREATE TABLE IF NOT EXISTS domain_bindings (
    id SERIAL PRIMARY KEY,
    domain TEXT UNIQUE NOT NULL,
    project_key TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    github_id TEXT UNIQUE NOT NULL,
    login TEXT,
    name TEXT,
    avatar_url TEXT,
    email TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_users_github_id ON users(github_id);

CREATE TABLE IF NOT EXISTS sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    session_token TEXT UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(session_token);
CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);

CREATE TABLE IF NOT EXISTS user_configs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER UNIQUE NOT NULL,
    config TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);


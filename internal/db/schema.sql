-- -------文章表--------
CREATE TABLE IF NOT EXISTS posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    slug TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    summary TEXT DEFAULT '',
    content_md TEXT NOT NULL,
    content_html TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'draft',
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    published_at DATETIME,
    render_version INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_posts_list ON posts(status,published_at DESC);

-- -------项目库表--------
CREATE TABLE IF NOT EXISTS projects (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    slug         TEXT     NOT NULL UNIQUE,
    title        TEXT     NOT NULL,
    summary      TEXT     NOT NULL,
    github_url   TEXT     NOT NULL,
    demo_url     TEXT     DEFAULT '',
    tech_stack   TEXT     DEFAULT '',
    cover_image  TEXT     DEFAULT '',
    status       TEXT     NOT NULL DEFAULT 'active',
    featured     INTEGER  NOT NULL DEFAULT 0,
    sort_order   INTEGER  NOT NULL DEFAULT 0,
    created_at   DATETIME NOT NULL,
    updated_at   DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_projects_order ON projects(featured DESC, sort_order ASC, created_at DESC);

-- ------管理员表-------

CREATE TABLE IF NOT EXISTS admins (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    username      TEXT NOT NULL UNIQUE,
    password_hash BLOB NOT NULL,
    salt          BLOB NOT NULL,
    iterations    INTEGER NOT NULL,
    created_at    DATETIME NOT NULL
);

-- ------会话表---------

CREATE TABLE IF NOT EXISTS sessions (
    token      TEXT PRIMARY KEY,
    admin_id   INTEGER NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (admin_id) REFERENCES admins(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);

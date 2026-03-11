-- 000001_init.up.sql

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Users
CREATE TABLE users (
    id                   UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name                 VARCHAR(100) NOT NULL,
    email                VARCHAR(255) NOT NULL UNIQUE,
    password_hash        TEXT        NOT NULL,
    profile              VARCHAR(20) NOT NULL CHECK (profile IN ('ADMIN', 'EXECUTOR', 'AUDITOR')),
    must_change_password BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);

-- Tasks
CREATE TABLE tasks (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    title       VARCHAR(200) NOT NULL,
    description TEXT         NOT NULL,
    status      VARCHAR(30)  NOT NULL DEFAULT 'ASSIGNED'
                    CHECK (status IN ('ASSIGNED','STARTED','FINISHED_SUCCESS','FINISHED_ERROR','ON_HOLD')),
    assignee_id UUID         NOT NULL REFERENCES users(id),
    due_date    TIMESTAMPTZ  NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tasks_assignee ON tasks(assignee_id);
CREATE INDEX idx_tasks_status   ON tasks(status);

-- Comments (only on expired tasks)
CREATE TABLE comments (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id    UUID        NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    author_id  UUID        NOT NULL REFERENCES users(id),
    body       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_comments_task ON comments(task_id);


INSERT INTO users (id, name, email, password_hash, profile, must_change_password)
VALUES (
    gen_random_uuid(),
    'System Administrator',
    'admin@taskmanager.local',
    '$2a$12$WvwOq2npSnIB7LxWB3KeBOdOl7itZGnKDvNlH9Mw2OuQl29HILi6.',
    'ADMIN',
    TRUE
);
INSERT INTO users (id, name, email, password_hash, profile, must_change_password)
VALUES (
    gen_random_uuid(),
    'Pedro Executor',
    'executor@taskmanager.local',
    '$2a$12$WvwOq2npSnIB7LxWB3KeBOdOl7itZGnKDvNlH9Mw2OuQl29HILi6.',
    'EXECUTOR',
    TRUE
);
INSERT INTO users (id, name, email, password_hash, profile, must_change_password)
VALUES (
    gen_random_uuid(),
    'Juan Auditor',
    'auditor@taskmanager.local',
    '$2a$12$WvwOq2npSnIB7LxWB3KeBOdOl7itZGnKDvNlH9Mw2OuQl29HILi6.',
    'AUDITOR',
    TRUE
);

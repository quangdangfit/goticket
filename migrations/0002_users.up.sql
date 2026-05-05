CREATE TABLE users (
    id              CHAR(26)        NOT NULL,
    email           VARCHAR(255)    NOT NULL,
    password_hash   VARCHAR(255)    NOT NULL,
    name            VARCHAR(255)    NOT NULL,
    phone           VARCHAR(32)     NOT NULL DEFAULT '',
    role            VARCHAR(16)     NOT NULL DEFAULT 'user',
    created_at      DATETIME(3)     NOT NULL,
    updated_at      DATETIME(3)     NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_users_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE refresh_tokens (
    id          CHAR(26)        NOT NULL,
    user_id     CHAR(26)        NOT NULL,
    token_hash  CHAR(64)        NOT NULL,
    expires_at  DATETIME(3)     NOT NULL,
    revoked_at  DATETIME(3)     NULL,
    created_at  DATETIME(3)     NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_refresh_tokens_hash (token_hash),
    KEY idx_refresh_tokens_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

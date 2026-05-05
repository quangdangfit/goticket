CREATE TABLE promo_codes (
    code            VARCHAR(64)     NOT NULL,
    type            VARCHAR(16)     NOT NULL,
    value_minor     BIGINT          NOT NULL,
    percent         INT             NOT NULL DEFAULT 0,
    max_uses        INT             NOT NULL,
    used            INT             NOT NULL DEFAULT 0,
    per_user_limit  INT             NOT NULL DEFAULT 1,
    starts_at       DATETIME(3)     NOT NULL,
    expires_at      DATETIME(3)     NOT NULL,
    created_at      DATETIME(3)     NOT NULL,
    PRIMARY KEY (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE promo_redemptions (
    id          CHAR(26)        NOT NULL,
    code        VARCHAR(64)     NOT NULL,
    user_id     CHAR(26)        NOT NULL,
    order_id    CHAR(26)        NOT NULL,
    discount_minor BIGINT       NOT NULL,
    created_at  DATETIME(3)     NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_promo_redemptions (code, order_id),
    KEY idx_promo_redemptions_user (code, user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

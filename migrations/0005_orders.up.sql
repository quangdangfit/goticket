CREATE TABLE orders (
    id              CHAR(26)        NOT NULL,
    user_id         CHAR(26)        NOT NULL,
    showtime_id     CHAR(26)        NOT NULL,
    hold_id         CHAR(26)        NOT NULL,
    status          VARCHAR(16)     NOT NULL DEFAULT 'pending',
    subtotal_minor  BIGINT          NOT NULL,
    discount_minor  BIGINT          NOT NULL DEFAULT 0,
    total_minor     BIGINT          NOT NULL,
    currency        CHAR(3)         NOT NULL,
    promo_code      VARCHAR(64)     NOT NULL DEFAULT '',
    created_at      DATETIME(3)     NOT NULL,
    updated_at      DATETIME(3)     NOT NULL,
    PRIMARY KEY (id),
    KEY idx_orders_user (user_id),
    KEY idx_orders_status (status),
    KEY idx_orders_hold (hold_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE order_items (
    id              CHAR(26)        NOT NULL,
    order_id        CHAR(26)        NOT NULL,
    ticket_type_id  CHAR(26)        NOT NULL,
    quantity        INT             NOT NULL,
    unit_price_minor BIGINT         NOT NULL,
    PRIMARY KEY (id),
    KEY idx_order_items_order (order_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE idempotency_keys (
    user_id     CHAR(26)        NOT NULL,
    `key`       VARCHAR(128)    NOT NULL,
    order_id    CHAR(26)        NOT NULL,
    created_at  DATETIME(3)     NOT NULL,
    PRIMARY KEY (user_id, `key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

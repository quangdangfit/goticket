CREATE TABLE payments (
    id              CHAR(26)        NOT NULL,
    order_id        CHAR(26)        NOT NULL,
    provider        VARCHAR(32)     NOT NULL,
    intent_id       VARCHAR(128)    NOT NULL,
    amount_minor    BIGINT          NOT NULL,
    currency        CHAR(3)         NOT NULL,
    status          VARCHAR(16)     NOT NULL,
    created_at      DATETIME(3)     NOT NULL,
    updated_at      DATETIME(3)     NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_payments_intent (provider, intent_id),
    KEY idx_payments_order (order_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE payment_events (
    id          CHAR(26)        NOT NULL,
    provider    VARCHAR(32)     NOT NULL,
    event_id    VARCHAR(128)    NOT NULL,
    intent_id   VARCHAR(128)    NOT NULL,
    type        VARCHAR(64)     NOT NULL,
    raw         MEDIUMTEXT      NOT NULL,
    received_at DATETIME(3)     NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_payment_events_event (provider, event_id),
    KEY idx_payment_events_intent (provider, intent_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

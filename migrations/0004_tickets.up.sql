CREATE TABLE ticket_types (
    id              CHAR(26)        NOT NULL,
    showtime_id     CHAR(26)        NOT NULL,
    name            VARCHAR(128)    NOT NULL,
    description     VARCHAR(512)    NOT NULL DEFAULT '',
    price_minor     BIGINT          NOT NULL,
    currency        CHAR(3)         NOT NULL DEFAULT 'VND',
    total_quota     INT             NOT NULL,
    per_user_limit  INT             NOT NULL DEFAULT 10,
    has_seat_map    TINYINT(1)      NOT NULL DEFAULT 0,
    created_at      DATETIME(3)     NOT NULL,
    updated_at      DATETIME(3)     NOT NULL,
    PRIMARY KEY (id),
    KEY idx_ticket_types_showtime (showtime_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE seats (
    id              CHAR(26)        NOT NULL,
    ticket_type_id  CHAR(26)        NOT NULL,
    section         VARCHAR(64)     NOT NULL,
    row_label       VARCHAR(16)     NOT NULL,
    seat_number     VARCHAR(16)     NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_seats_position (ticket_type_id, section, row_label, seat_number),
    KEY idx_seats_type (ticket_type_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

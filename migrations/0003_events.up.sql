CREATE TABLE venues (
    id          CHAR(26)        NOT NULL,
    name        VARCHAR(255)    NOT NULL,
    address     VARCHAR(512)    NOT NULL,
    city        VARCHAR(128)    NOT NULL,
    capacity    INT             NOT NULL,
    created_at  DATETIME(3)     NOT NULL,
    updated_at  DATETIME(3)     NOT NULL,
    PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE events (
    id          CHAR(26)        NOT NULL,
    title       VARCHAR(255)    NOT NULL,
    description TEXT            NOT NULL,
    organizer   VARCHAR(255)    NOT NULL,
    poster_url  VARCHAR(1024)   NOT NULL DEFAULT '',
    status      VARCHAR(16)     NOT NULL DEFAULT 'draft',
    venue_id    CHAR(26)        NOT NULL,
    created_at  DATETIME(3)     NOT NULL,
    updated_at  DATETIME(3)     NOT NULL,
    PRIMARY KEY (id),
    KEY idx_events_status (status),
    KEY idx_events_venue (venue_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE showtimes (
    id              CHAR(26)        NOT NULL,
    event_id        CHAR(26)        NOT NULL,
    starts_at       DATETIME(3)     NOT NULL,
    ends_at         DATETIME(3)     NOT NULL,
    sales_open_at   DATETIME(3)     NOT NULL,
    sales_close_at  DATETIME(3)     NOT NULL,
    status          VARCHAR(16)     NOT NULL DEFAULT 'scheduled',
    created_at      DATETIME(3)     NOT NULL,
    updated_at      DATETIME(3)     NOT NULL,
    PRIMARY KEY (id),
    KEY idx_showtimes_event (event_id),
    KEY idx_showtimes_starts (starts_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

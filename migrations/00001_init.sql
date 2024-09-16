-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';



-- Создание таблицы tender
CREATE TABLE tender
(
    id                 UUID PRIMARY KEY,
    organization_id    UUID         NOT NULL,
    status             VARCHAR(255) NOT NULL,
    created_at         TIMESTAMP    NOT NULL DEFAULT NOW(),
    current_version_id UUID
);

-- Индекс на поле created_at
CREATE INDEX idx_tender_created_at ON tender (created_at);

-- Создание таблицы tender_version
CREATE TABLE tender_version
(
    id           UUID PRIMARY KEY,
    tender_id    UUID         NOT NULL,
    version      INT          NOT NULL,
    created_at   TIMESTAMP    NOT NULL DEFAULT NOW(),
    name         VARCHAR(255) NOT NULL,
    description  TEXT         NOT NULL,
    service_type VARCHAR(255) NOT NULL,
    CONSTRAINT fk_tender FOREIGN KEY (tender_id) REFERENCES tender (id)
);

-- Индекс на поле created_at
CREATE INDEX idx_tender_version_created_at ON tender_version (created_at);

-- Создание таблицы bid
CREATE TABLE bid
(
    id                 UUID PRIMARY KEY,
    tender_id          UUID         NOT NULL,
    status             VARCHAR(255) NOT NULL,
    author_type        VARCHAR(255) NOT NULL,
    author_id          UUID         NOT NULL,
    current_version_id UUID,
    created_at         TIMESTAMP    NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_tender FOREIGN KEY (tender_id) REFERENCES tender (id)
);

-- Создание таблицы bid_version
CREATE TABLE bid_version
(
    id          UUID PRIMARY KEY,
    bid_id      UUID         NOT NULL,
    version     INT          NOT NULL,
    created_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
    name        VARCHAR(255) NOT NULL,
    description TEXT         NOT NULL,
    CONSTRAINT fk_bid FOREIGN KEY (bid_id) REFERENCES bid (id)
);

-- Индекс на поле created_at
CREATE INDEX idx_bid_version_created_at ON bid_version (created_at);

-- Создание таблицы bid_decision
CREATE TABLE bid_decision
(
    id          UUID PRIMARY KEY,
    bid_id      UUID         NOT NULL,
    decision    VARCHAR(255) NOT NULL,
    employee_id UUID         NOT NULL,
    created_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_bid FOREIGN KEY (bid_id) REFERENCES bid (id)
);

-- Индекс на поле created_at
CREATE INDEX idx_bid_decision_created_at ON bid_decision (created_at);

-- Создание таблицы bid_feedback
CREATE TABLE bid_feedback
(
    id          UUID PRIMARY KEY,
    bid_id      UUID      NOT NULL,
    description TEXT      NOT NULL,
    author_id   UUID      NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_bid FOREIGN KEY (bid_id) REFERENCES bid (id)
);

-- Индекс на поле created_at
CREATE INDEX idx_bid_feedback_created_at ON bid_feedback (created_at);

CREATE TABLE employee
(
    id         UUID PRIMARY KEY,
    username   VARCHAR(50) UNIQUE NOT NULL,
    first_name VARCHAR(50),
    last_name  VARCHAR(50),
    created_at TIMESTAMP        DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP        DEFAULT CURRENT_TIMESTAMP
);


CREATE TYPE organization_type AS ENUM (
    'IE',
    'LLC',
    'JSC'
    );

CREATE TABLE organization
(
    id          UUID PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    type        organization_type,
    created_at  TIMESTAMP        DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP        DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE organization_responsible
(
    id              UUID PRIMARY KEY,
    organization_id UUID REFERENCES organization (id) ON DELETE CASCADE,
    user_id         UUID REFERENCES employee (id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

DROP TABLE bid_decision CASCADE;
DROP TABLE bid_feedback CASCADE;
DROP TABLE bid_version CASCADE;
DROP TABLE bid CASCADE;
DROP TABLE tender_version CASCADE;
DROP TABLE tender CASCADE;
DROP TABLE employee CASCADE;
DROP TABLE organization_responsible CASCADE;
DROP TABLE organization CASCADE;
DROP TYPE organization_type;
-- +goose StatementEnd

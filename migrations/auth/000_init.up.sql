CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE TABLE IF NOT EXISTS app_sessions
(
    id          UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    fingerprint JSONB       NOT NULL,
    issued_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at  TIMESTAMPTZ NOT NULL,
    jwt_id      UUID        NOT NULL,
    revoked     BOOLEAN     NOT NULL DEFAULT FALSE
);
CREATE INDEX idx_app_sessions_device_id
    ON app_sessions ((fingerprint ->> 'device_id'));
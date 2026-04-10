CREATE TABLE IF NOT EXISTS pastes (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    public_id   VARCHAR(50)  NOT NULL,
    title       VARCHAR(255) NOT NULL,
    content     TEXT         NOT NULL,
    size        INT          NOT NULL,
    language    VARCHAR(50)  NOT NULL,
    ip_hash     VARCHAR(255) NOT NULL,
    user_agent  VARCHAR(512),
    expires_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS pastes_public_id_idx ON pastes (public_id);

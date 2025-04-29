CREATE TABLE IF NOT EXISTS manifest
(
    id              UUID PRIMARY KEY,
    version         TEXT        NOT NULL, -- "1.0.0"
    icon            TEXT        NOT NULL, -- lock.shield
    category        TEXT        NOT NULL, -- security
    tags            TEXT[]      NOT NULL, -- ARRAY['rsa','encryption','crypto']
    author_name     TEXT        NOT NULL, -- Sergey K.
    author_email    TEXT        NOT NULL, -- sergey@example.com
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    meta_created_at TIMESTAMPTZ NOT NULL, -- из JSON.meta.createdAt
    signature       TEXT        NOT NULL  -- Base64-Ed25519 от ui+script
);

CREATE INDEX IF NOT EXISTS idx_manifest_category
    ON manifest (category);

CREATE INDEX IF NOT EXISTS idx_manifest_tags
    ON manifest USING GIN (tags);

CREATE TABLE IF NOT EXISTS manifest_localizations
(
    manifest_id UUID NOT NULL REFERENCES manifest (id) ON DELETE CASCADE,
    locale      TEXT NOT NULL,
    key         TEXT NOT NULL,
    value       TEXT NOT NULL,
    PRIMARY KEY (manifest_id, locale, key)
);

CREATE INDEX IF NOT EXISTS idx_key_manifest_locale
    ON manifest_localizations (manifest_id, locale);


CREATE TABLE IF NOT EXISTS manifest_content
(
    manifest_id UUID PRIMARY KEY REFERENCES manifest (id) ON DELETE CASCADE,
    ui          JSONB  NOT NULL, -- { layout:…, components:… }
    script      TEXT   NOT NULL, -- весь JS-код
    actions     JSONB  NOT NULL, -- [ { id, label, icon, onTap }, … ]
    permissions TEXT[] NOT NULL  -- ARRAY['clipboard.write', 'share', …]
);

CREATE INDEX IF NOT EXISTS idx_manifest_content_ui
    ON manifest_content USING GIN (ui);

CREATE INDEX IF NOT EXISTS idx_manifest_content_actions
    ON manifest_content USING GIN (actions);

CREATE INDEX IF NOT EXISTS idx_manifest_content_permissions
    ON manifest_content USING GIN (permissions);
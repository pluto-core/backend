-- name: CreateManifest :one
INSERT INTO manifest (id, version, icon, category, tags,
                      author_name, author_email,
                      meta_created_at)
VALUES ($1, $2, $3, $4, $5,
        $6, $7,
        $8)
RETURNING
    id, version, icon, category, tags,
    author_name, author_email,
    created_at, meta_created_at;


-- name: GetManifestByID :one
SELECT id,
       version,
       icon,
       category,
       tags,
       author_name,
       author_email,
       created_at,
       meta_created_at
FROM manifest
WHERE id = $1;


-- name: ListManifests :many
SELECT id,
       version,
       icon,
       category,
       tags,
       author_name,
       author_email,
       created_at,
       meta_created_at
FROM manifest
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;


-- name: DeleteManifestByID :exec
DELETE
FROM manifest
WHERE id = $1;

-- name: CreateManifestContent :one
INSERT INTO manifest_content (manifest_id, ui, script, actions, permissions, signature)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING manifest_id, ui, script, actions, permissions, signature;


-- name: GetManifestContentByID :one
SELECT manifest_id,
       ui,
       script,
       actions,
       permissions,
       signature
FROM manifest_content
WHERE manifest_id = $1;


-- name: UpdateManifestContent :one
UPDATE manifest_content
SET ui          = $2,
    script      = $3,
    actions     = $4,
    permissions = $5,
    signature   = $6
WHERE manifest_id = $1
RETURNING manifest_id, ui, script, actions, permissions, signature;


-- name: DeleteManifestContentByID :exec
DELETE
FROM manifest_content
WHERE manifest_id = $1;

-- name: SearchManifests :many
SELECT m.id,
       m.version,
       m.icon,
       m.category,
       m.tags,
       m.author_name,
       m.author_email,
       m.created_at,
       m.meta_created_at
FROM manifest m
         LEFT JOIN manifest_localizations t
                   ON t.manifest_id = m.id
                       AND t.locale = $2::text
                       AND t.key = 'title'
         LEFT JOIN manifest_localizations d
                   ON d.manifest_id = m.id
                       AND d.locale = $2::text
                       AND d.key = 'description'
WHERE (
          -- empty search string => return everything
          ($1::text = '')
              -- full-text on title+description
              OR to_tsvector('english',
                             coalesce(t.value, '') || ' ' || coalesce(d.value, '')
                 ) @@ plainto_tsquery('english', $1::text)
              -- category ilike
              OR m.category ILIKE '%' || $1::text || '%'
              -- any tag matches
              OR EXISTS (SELECT 1
                         FROM unnest(m.tags) AS tag
                         WHERE tag ILIKE '%' || $1::text || '%')
          )
ORDER BY m.created_at DESC
LIMIT $3::int OFFSET $4::int;

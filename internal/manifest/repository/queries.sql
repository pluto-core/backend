-- name: ListManifests :many
SELECT m.id,
       m.version,
       m.icon,
       m.category,
       m.tags,
       m.author_name,
       m.author_email,
       m.created_at,
       m.meta_created_at,
       -- 🔽 отдаём **всю** локализацию для запрошенной локали
       (SELECT jsonb_object_agg(l.key, l.value)
        FROM manifest_localizations AS l
        WHERE l.manifest_id = m.id
          AND l.locale = $3::text) AS localization
FROM manifest AS m
ORDER BY m.created_at DESC
LIMIT $1 OFFSET $2;


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
                       AND t.locale = sqlc.arg(locale)
                       AND t.key = 'title'
         LEFT JOIN manifest_localizations d
                   ON d.manifest_id = m.id
                       AND d.locale = sqlc.arg(locale)
                       AND d.key = 'description'
WHERE (
          -- empty search string => return everything
          (sqlc.arg(search)::text = '')
              -- full-text on title+description
              OR to_tsvector('english',
                             coalesce(t.value, '') || ' ' || coalesce(d.value, '')
                 ) @@ plainto_tsquery('english', sqlc.arg(search)::text)
              -- category ilike
              OR m.category ILIKE '%' || sqlc.arg(search)::text || '%'
              -- any tag matches
              OR EXISTS (SELECT 1
                         FROM unnest(m.tags) AS tag
                         WHERE tag ILIKE '%' || sqlc.arg(search)::text || '%')
          )
ORDER BY m.created_at DESC;


-- name: SearchManifestsFTS :many
SELECT m.id,
       m.version,
       m.icon,
       m.category,
       m.tags,
       m.author_name,
       m.author_email,
       m.created_at,
       m.meta_created_at,
       (SELECT jsonb_object_agg(l.key, l.value)
        FROM manifest_localizations AS l
        WHERE l.manifest_id = m.id
          AND l.locale = sqlc.arg(locale)::text) AS localization
FROM manifest AS m
WHERE to_tsvector(sqlc.arg(config)::regconfig,
                  (SELECT string_agg(l.value, ' ')
                   FROM manifest_localizations AS l
                   WHERE l.manifest_id = m.id
                     AND l.locale = sqlc.arg(locale)::text
                     AND l.key IN ('title', 'description')) || ' ' ||
                  m.category || ' ' ||
                  array_to_string(m.tags, ' ')
      )
          @@ plainto_tsquery(sqlc.arg(config)::regconfig, sqlc.arg(query)::text)
ORDER BY m.created_at DESC;

-- name: GetManifest :one
WITH localization AS (SELECT ml.manifest_id,
                             json_object_agg(ml.key, ml.value) AS localization
                      FROM manifest_localizations ml
                      WHERE ml.locale = sqlc.arg(locale)::text
                      GROUP BY ml.manifest_id)
SELECT m.id,
       m.version,
       m.icon,
       m.category,
       m.tags,
       m.author_name,
       m.author_email,
       m.created_at,
       m.meta_created_at,
       m.signature,
       mc.ui AS U_I,
       mc.script,
       mc.actions,
       mc.permissions,
       l.localization
FROM manifest m
         LEFT JOIN manifest_content mc ON mc.manifest_id = m.id
         LEFT JOIN localization l ON l.manifest_id = m.id
WHERE m.id = sqlc.arg(manifest_id)::uuid;

-- name: CreateManifest :one
INSERT INTO manifest (id,
                      version,
                      icon,
                      category,
                      tags,
                      author_name,
                      author_email,
                      created_at,
                      meta_created_at,
                      signature)
VALUES (sqlc.arg(id),
        sqlc.arg(version),
        sqlc.arg(icon),
        sqlc.arg(category),
        sqlc.arg(tags),
        sqlc.arg(author_name),
        sqlc.arg(author_email),
        now(),
        now(),
        sqlc.arg(signature))
RETURNING id;

-- name: CreateManifestContent :exec
INSERT INTO manifest_content (manifest_id,
                              ui,
                              script,
                              actions,
                              permissions)
VALUES (sqlc.arg(manifest_id),
        sqlc.arg(ui),
        sqlc.arg(script),
        sqlc.arg(actions),
        sqlc.arg(permissions));

-- name: CreateLocalizations :exec
INSERT INTO manifest_localizations (manifest_id,
                                    locale,
                                    key,
                                    value)
SELECT sqlc.arg(manifest_id),
       unnest(sqlc.arg(locales)::text[]),
       unnest(sqlc.arg(keys)::text[]),
       unnest(sqlc.arg(values)::text[]);
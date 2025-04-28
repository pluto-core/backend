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
       t.value AS title,
       d.value AS description
FROM manifest m
         LEFT JOIN manifest_localizations t
                   ON t.manifest_id = m.id
                       AND t.locale = $3::text
                       AND t.key = 'title'
         LEFT JOIN manifest_localizations d
                   ON d.manifest_id = m.id
                       AND d.locale = $3::text
                       AND d.key = 'description'
ORDER BY created_at DESC
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
       t.value AS title,
       d.value AS description,
       m.version,
       m.icon,
       m.category,
       m.tags,
       m.author_name,
       m.author_email,
       m.created_at,
       m.meta_created_at
FROM manifest AS m
         LEFT JOIN manifest_localizations AS t
                   ON t.manifest_id = m.id AND t.locale = sqlc.arg(locale)::text AND t.key = 'title'
         LEFT JOIN manifest_localizations AS d
                   ON d.manifest_id = m.id AND d.locale = sqlc.arg(locale)::text AND d.key = 'description'
WHERE to_tsvector(sqlc.arg(config)::regconfig,
                  coalesce(t.value, '') || ' ' ||
                  coalesce(d.value, '') || ' ' ||
                  m.category || ' ' ||
                  array_to_string(m.tags, ' ')
      )
          @@ plainto_tsquery(sqlc.arg(config)::regconfig, sqlc.arg(query)::text)
ORDER BY m.created_at DESC;
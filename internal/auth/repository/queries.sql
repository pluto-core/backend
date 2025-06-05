-- name: GetActiveAppSessionByDeviceID :one
SELECT id, jwt_id, expires_at
FROM app_sessions
WHERE (fingerprint ->>'device_id') = $1
  AND revoked = false
  AND expires_at > now();

-- name: CreateAppSession :exec
INSERT INTO app_sessions (id, fingerprint, issued_at, expires_at, jwt_id, revoked)
VALUES ($1, $2, now(), $3, $4, false);

-- name: RevokeAppSession :exec
UPDATE app_sessions
SET revoked = true
WHERE id = $1
  AND jwt_id = $2;
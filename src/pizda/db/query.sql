-- name: GetUsers :many
SELECT * FROM pizda.user WHERE is_blocked = false;

-- name: UpsertUser :exec
INSERT INTO pizda.user (tg_id, username, first_name, last_name) VALUES ($1, $2, $3, $4)
  ON CONFLICT(tg_id) DO UPDATE SET
  username = EXCLUDED.username,
  first_name = EXCLUDED.first_name,
  last_name = EXCLUDED.last_name,
  is_blocked = false;

-- name: GetUsersIDs :many
SELECT tg_id FROM pizda.user WHERE is_blocked = false;

-- name: BlockUsers :exec
UPDATE pizda.user SET is_blocked = true WHERE tg_id = ANY(sqlc.arg(ids)::bigint[]);

-- name: IfUserPays :many
SELECT * FROM pizda.payment WHERE user_id = $1 and period @> NOW();

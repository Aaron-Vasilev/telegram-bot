-- name: GetUsers :many
SELECT * FROM pizda.user WHERE is_blocked = false;

-- name: FindUsersByName :many
SELECT * FROM pizda.user WHERE first_name ILIKE '%' || @name::text || '%' OR last_name ILIKE '%' || @name::text || '%' OR username ILIKE '%' || @name::text || '%';

-- name: UpsertUser :exec
INSERT INTO pizda.user (id, username, first_name, last_name) VALUES ($1, $2, $3, $4)
  ON CONFLICT(id) DO UPDATE SET
  username = EXCLUDED.username,
  first_name = EXCLUDED.first_name,
  last_name = EXCLUDED.last_name,
  is_blocked = false;

-- name: GetUsersIDs :many
SELECT id FROM pizda.user WHERE is_blocked = false;

-- name: BlockUsers :exec
UPDATE pizda.user SET is_blocked = true WHERE id = ANY(sqlc.arg(ids)::bigint[]);

-- name: GetValidPayment :one
SELECT * FROM pizda.payment WHERE user_id = $1 and period @> NOW()::DATE limit 1;

-- name: AddPayment :exec
INSERT INTO pizda.payment (user_id, method) VALUES ($1, $2);

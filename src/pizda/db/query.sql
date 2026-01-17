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

-- name: GetVideos :many
SELECT * FROM pizda.file order by id;

-- name: GetVideoById :one
SELECT * FROM pizda.file WHERE id=$1;

-- name: UpdateFileId :exec
UPDATE pizda.file SET file_id=$1 WHERE id=$2;

-- name: GetPaymentsEndingSoon :many
SELECT p.*, u.id as user_id, u.first_name, u.last_name, u.username
FROM pizda.payment p
JOIN pizda.user u ON p.user_id = u.id
WHERE upper(p.period) - CURRENT_DATE <= 3
  AND upper(p.period) - CURRENT_DATE >= 0
  AND p.is_notified = false;

-- name: MarkPaymentAsNotified :exec
UPDATE pizda.payment SET is_notified = true WHERE id = $1;

-- name: ExtendPaymentByMonth :exec
UPDATE pizda.payment
SET period = daterange(lower(period), (upper(period) + INTERVAL '1 month')::date, '[]')
WHERE id = $1;

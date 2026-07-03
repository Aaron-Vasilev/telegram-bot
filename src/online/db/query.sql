-- name: UpsertUser :exec
INSERT INTO online."user" (id, username, first_name, last_name)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id) DO UPDATE
SET username = EXCLUDED.username,
    first_name = EXCLUDED.first_name,
    last_name = EXCLUDED.last_name;

-- name: GetUser :one
SELECT * FROM online."user" WHERE id = $1;

-- name: FindUsersByName :many
SELECT * FROM online."user"
WHERE first_name ILIKE '%' || $1 || '%'
   OR last_name ILIKE '%' || $1 || '%'
   OR username ILIKE '%' || $1 || '%';

-- name: GetActiveSubscription :one
SELECT * FROM online.subscription
WHERE user_id = $1 AND ends >= CURRENT_DATE
ORDER BY ends DESC
LIMIT 1;

-- name: GetSubscriptionByPaymentRef :one
SELECT * FROM online.subscription
WHERE payment_ref = $1
ORDER BY id DESC
LIMIT 1;

-- name: CreateSubscription :one
INSERT INTO online.subscription (user_id, payment_ref, payment_token, starts, ends, is_manual)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ExtendSubscription :exec
UPDATE online.subscription
SET ends = $2, is_notified = false
WHERE id = $1;

-- name: MarkSubscriptionNotified :exec
UPDATE online.subscription
SET is_notified = true
WHERE id = $1;

-- name: DeactivateSubscription :exec
UPDATE online.subscription
SET is_active = false
WHERE id = $1;

-- name: GetSubscriptionsForRenewal :many
SELECT s.*, u.username, u.first_name, u.last_name
FROM online.subscription s
JOIN online."user" u ON s.user_id = u.id
WHERE s.ends = CURRENT_DATE
  AND s.is_active = true
  AND s.is_manual = false
  AND s.payment_token <> '';

-- name: GetManualSubscriptionsEndingTomorrow :many
SELECT s.*, u.username, u.first_name, u.last_name
FROM online.subscription s
JOIN online."user" u ON s.user_id = u.id
WHERE s.ends = CURRENT_DATE + 1
  AND s.is_notified = false
  AND s.is_manual = true;

-- name: GetExpiredUsers :many
SELECT u.id
FROM online."user" u
WHERE u.is_blocked = false
  AND NOT EXISTS (
    SELECT 1 FROM online.subscription s
    WHERE s.user_id = u.id AND s.ends >= CURRENT_DATE
  );
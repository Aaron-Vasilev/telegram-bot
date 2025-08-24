-- name: GetUsersIDs :many
SELECT id FROM yoga.user WHERE is_blocked = false;

-- name: FindUsersByName :many
SELECT * FROM yoga.user WHERE name ILIKE '%' || @name::text || '%' OR username ILIKE '%' || @name::text || '%';

-- name: UpdateEmoji :exec
UPDATE yoga.user SET emoji=$1 WHERE id=$2;

-- name: UpsertUser :exec
INSERT INTO yoga.user (id, username, name) VALUES ($1, $2, $3)
  ON CONFLICT(id) DO UPDATE SET
  username = EXCLUDED.username, name = EXCLUDED.name, is_blocked = false;

-- name: GetUserWithMembership :one
SELECT u.id, username, name, emoji, starts, ends, type, lessons_avaliable
FROM yoga.user u LEFT JOIN yoga.membership m ON u.id = m.user_id WHERE u.id=$1;

-- name: GetAllUsersWithMemLatest :many
SELECT u.id, username, name, emoji, starts, ends, type, lessons_avaliable
FROM yoga.user u INNER JOIN yoga.membership m ON u.id = m.user_id 
WHERE m.ends >= NOW() - INTERVAL '2 months' AND is_blocked = false;

-- name: GetAvailableLessons :many
SELECT * FROM yoga.lesson WHERE (date >= (now())::date);

-- name: UpdateUserBio :exec
UPDATE yoga.user SET username=$1, name=$2 WHERE id=$3;

-- name: GetLessonWithUsers :many
SELECT
  l.id as lesson_id,
  date,
  time,
  max,
  description,
  u.id as user_id,
  u.name as name,
  u.username as username,
  u.emoji as emoji
FROM yoga.lesson l LEFT JOIN yoga.registered_users r ON l.id = r.lesson_id
LEFT JOIN yoga.user u ON u.id = ANY(r.registered) WHERE l.id=$1;

-- name: RegisterUser :exec
UPDATE yoga.registered_users
SET registered = array_append(registered, @user_id)
WHERE lesson_id=$1 AND NOT (@user_id=ANY(registered));

-- name: UnregisterUser :exec
UPDATE yoga.registered_users
SET registered = array_remove(registered, $1)
WHERE lesson_id=$2;

-- name: GetRegisteredUsers :many
SELECT * FROM yoga.registered_users WHERE lesson_id=$1;

-- name: GetRegisteredOnLesson :one
SELECT registered, lesson_id, date
FROM yoga.registered_users INNER JOIN yoga.lesson on id=lesson_id 
WHERE lesson_id=$1;

-- name: AddRegisterdUsersRow :exec
INSERT INTO yoga.registered_users (lesson_id, registered) VALUES ($1, $2);

-- name: BlockUsers :exec
UPDATE yoga.user SET is_blocked = true WHERE id = ANY(sqlc.arg(ids)::bigint[]);

-- name: GetLessonsByDate :many
SELECT * FROM yoga.lesson WHERE date=$1;

-- name: UpdateLessonDate :exec
UPDATE yoga.lesson SET date=$1 WHERE id=$2;

-- name: UpdateLessonTime :exec
UPDATE yoga.lesson SET time=$1 WHERE id=$2;

-- name: UpdateLessonDesc :exec
UPDATE yoga.lesson SET description=$1 WHERE id=$2;

-- name: UpdateLessonMax :exec
UPDATE yoga.lesson SET max=$1 WHERE id=$2;

-- name: AddLesson :one
INSERT INTO yoga.lesson (date, time, description, max) 
VALUES ($1, $2, $3, $4) RETURNING id;

-- name: GetFirstLessonAttendanceId :one
SELECT id FROM yoga.attendance WHERE lesson_id=$1 limit 1;

-- name: GetUsersAttandance :many
SELECT u.id, username, name, emoji, COUNT(user_id) AS count FROM yoga.attendance 
JOIN yoga.user u ON u.id = user_id WHERE date>= @from_date and date <= @to_date 
GROUP BY u.id, username, name, emoji ORDER BY count DESC;

-- name: AddAttendance :exec
INSERT INTO yoga.attendance (user_id, lesson_id, date) VALUES ($1, $2, $3);

-- name: AddDaysToMem :exec
UPDATE yoga.membership SET ends = ends + $1 * INTERVAL '1 days' WHERE user_id=$2;

-- name: GetUsersIDsWithValidMem :many
SELECT user_id FROM yoga.membership m WHERE m.ends > NOW();

-- name: GetMembership :one
SELECT * FROM yoga.membership WHERE user_id=$1;

-- name: CheckIfUserHasCourseAccess :many
SELECT user_id FROM yoga.subscription WHERE user_id=$1 AND ends >= NOW();

-- name: UpdateMembership :exec
INSERT INTO yoga.membership (user_id, starts, ends, type, lessons_avaliable)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (user_id) -- This is the column that might conflict
DO UPDATE SET
  type = EXCLUDED.type,
  starts = EXCLUDED.starts,
  ends = EXCLUDED.ends,
  lessons_avaliable = EXCLUDED.lessons_avaliable;

-- name: DecLessonsAvaliable :exec
UPDATE yoga.membership SET lessons_avaliable = lessons_avaliable - 1 WHERE user_id=$1;

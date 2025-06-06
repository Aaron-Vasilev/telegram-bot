package controller

import (
	"bot/src/utils"
	t "bot/src/utils/types"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
)

func GetAvaliableLessons(db *sql.DB) []t.Lesson {
	var lessons []t.Lesson
	query := "SELECT * FROM yoga.available_lessons ORDER BY date ASC;"

	rows, err := db.Query(query)

	if err == nil {
		for rows.Next() {
			var l t.Lesson

			err := rows.Scan(&l.ID, &l.Date, &l.Time, &l.Description, &l.Max)

			if err != nil {
				fmt.Println("✡️  line 22 err", err)
			}

			lessons = append(lessons, l)
		}
	}
	defer rows.Close()

	return lessons
}

func GetLessonWithUsers(db *sql.DB, callBackData string) (t.LessonWithUsers, error) {
	var lesson t.LessonWithUsers
	query := `SELECT u.id, username, name, emoji, l.id, time, date, description, max
		FROM yoga.lesson l LEFT JOIN yoga.registered_users r ON l.id = r.lesson_id
		LEFT JOIN yoga.user u ON u.id = ANY(r.registered) WHERE l.id=$1;`

	data := strings.Split(callBackData, "=")

	rows, err := db.Query(query, data[1])

	if err == nil {
		for rows.Next() {
			var userId sql.NullInt64
			var username sql.NullString
			var name sql.NullString
			var emoji sql.NullString

			err := rows.Scan(
				&userId, &username, &name, &emoji,
				&lesson.LessonId, &lesson.Time, &lesson.Date, &lesson.Description, &lesson.Max,
			)

			if err != nil {
				return lesson, err
			} else if userId.Valid {
				lesson.Users = append(lesson.Users, t.UserDB{
					ID:       userId.Int64,
					Username: username.String,
					Name:     name.String,
					Emoji:    emoji.String,
				})
			}
		}
	}
	defer rows.Close()

	return lesson, nil
}

func ToggleUserInLesson(db *sql.DB, u t.Update) string {
	text := ""
	data := strings.Split(u.CallbackQuery.Data, "=")
	action := data[0]
	lessonId := data[1]

	switch action {
	case "REGISTER":
		query := `UPDATE yoga.registered_users
			SET registered = array_append(registered, $1)
			WHERE lesson_id=$2 AND NOT ($1=ANY(registered));`

		db.Exec(query, u.CallbackQuery.From.ID, lessonId)

		text = utils.SeeYouMsg
	case "UNREGISTER":
		query := `UPDATE yoga.registered_users
			SET registered = array_remove(registered, $1)
			WHERE lesson_id=$2;`

		db.Exec(query, u.CallbackQuery.From.ID, lessonId)

		text = "You are free, fatass...🌚"
	}

	return text
}

func CreateLesson() {
	//TODO don't forget to add a row into registered_users
}

func SaveUser(db *sql.DB, id int64, username string, name string) {
	query := `
	INSERT INTO yoga.user (id, username, name) VALUES ($1, $2, $3)
	ON CONFLICT(id) DO UPDATE SET
	username = EXCLUDED.username, name = EXCLUDED.name, is_blocked = false;`

	db.Query(query, id, username, name)
	//Notification for a Teacher if the user is new
}

func GetUserWithMembership(db *sql.DB, userId int64) t.UserMembership {
	var u t.UserMembership
	query := `
	SELECT u.id, username, name, emoji, starts, ends, type, lessons_avaliable
	FROM yoga.user u LEFT JOIN yoga.membership m ON u.id = m.user_id WHERE u.id=$1;`

	db.QueryRow(query, userId).Scan(
		&u.User.ID, &u.User.Username, &u.User.Name, &u.User.Emoji,
		&u.Starts, &u.Ends, &u.Type, &u.LessonsAvailable,
	)

	return u
}

func GetAllUsersWithMemLatest(db *sql.DB) []t.UserMembership {
	var userMem []t.UserMembership
	query := `
	SELECT u.id, username, name, emoji, starts, ends, type, lessons_avaliable
	FROM yoga.user u LEFT JOIN yoga.membership m ON u.id = m.user_id 
	WHERE m.ends >= NOW() - INTERVAL '2 months' AND is_blocked = false;`

	rows, err := db.Query(query)

	if err == nil {
		for rows.Next() {
			var u t.UserMembership

			err := rows.Scan(
				&u.User.ID, &u.User.Username, &u.User.Name, &u.User.Emoji,
				&u.Starts, &u.Ends, &u.Type, &u.LessonsAvailable,
			)

			if err != nil {
				fmt.Println("✡️  line 233 err", err)
			}

			userMem = append(userMem, u)
		}
	}

	return userMem
}

func UpdateEmoji(db *sql.DB, userId int64, emoji string) {
	query := `UPDATE yoga.user SET emoji=$1 WHERE id=$2`

	db.Exec(query, emoji, userId)
}

func AddLesson(db *sql.DB, l utils.ValidatedLesson) {
	query := `INSERT INTO yoga.lesson (date, time, description, max) 
	VALUES ($1, $2, $3, $4) RETURNING id;`
	var id int

	db.QueryRow(query, l.Date, l.Time, l.Description, l.Max).Scan(&id)

	query = `INSERT INTO yoga.registered_users (lesson_id, registered) VALUES ($1, $2);`

	db.Exec(query, id, pq.Array([]int{}))
}

func FindUsersByName(db *sql.DB, name string) ([]t.UserDB, error) {
	var users []t.UserDB
	query := `SELECT * FROM yoga.user WHERE name ILIKE '%' || $1 || '%' OR username ILIKE '%' || $1 || '%';`

	rows, err := db.Query(query, name)

	if err != nil {
		return users, err
	}

	for rows.Next() {
		var u t.UserDB
		var username sql.NullString

		err := rows.Scan(&u.ID, &username, &u.Name, &u.Emoji, &u.IsBlocked)

		if err != nil {
			return users, nil
		}

		if username.Valid {
			u.Username = username.String
		}

		users = append(users, u)
	}
	defer rows.Close()

	return users, nil
}

func UpdateMembership(db *sql.DB, userId int64, memType int) t.Membership {
	var m t.Membership
	token := t.Token{
		Created: time.Now(),
		Type:    memType,
	}

	query := `SELECT * FROM yoga.membership WHERE user_id=$1;`
	err := db.QueryRow(query, userId).Scan(&m.UserID, &m.Starts, &m.Ends, &m.Type, &m.LessonsAvailable)

	if err == sql.ErrNoRows {
		tmp := "2023-01-01"
		db.Exec(`INSERT INTO yoga.membership (user_id, starts, ends, type, lessons_avaliable)
		VALUES ($1,$2,$3,$4,$5);`, userId, tmp, tmp, 0, 0)
	}

	utils.UpdateMembership(&m, token)

	query = `UPDATE yoga.membership 
	SET type=$1, starts=$2, ends=$3, lessons_avaliable=$4 WHERE user_id=$5;`

	db.QueryRow(query, m.Type, m.Starts, m.Ends, m.LessonsAvailable, userId)

	return m
}

func GetRegisteredOnLesson(db *sql.DB, lessonId int) t.RegisterdOnLesson {
	var r t.RegisterdOnLesson
	query := `SELECT registered, lesson_id, date
FROM yoga.registered_users JOIN yoga.lesson on id=lesson_id 
WHERE lesson_id=$1;`

	db.QueryRow(query, lessonId).Scan(pq.Array(&r.IDs), &r.LessonId, &r.Date)

	return r
}

func GetUsersIDs(db *sql.DB) []int64 {
	var ids []int64

	rows, err := db.Query("SELECT id FROM yoga.user WHERE is_blocked = false;")
	defer rows.Close()

	if err == nil {
		for rows.Next() {
			var id int64

			err := rows.Scan(&id)

			if err != nil {
				fmt.Println("✡️  line 233 err", err)
			}

			ids = append(ids, id)
		}
	}

	return ids
}

func GetUsersIDsWithValidMem(db *sql.DB) []int64 {
	var ids []int64

	rows, err := db.Query("SELECT user_id FROM yoga.membership m WHERE m.ends > NOW();")
	defer rows.Close()

	if err == nil {
		for rows.Next() {
			var id int64

			err := rows.Scan(&id)

			if err != nil {
				fmt.Println("✡️  line 268 err", err)
			}

			ids = append(ids, id)
		}
	}

	return ids
}

func AddDaysToMem(db *sql.DB, userId int64, num int) {
	query := "UPDATE yoga.membership SET ends = ends + $1 * INTERVAL '1 days' WHERE user_id=$2;"
	_, err := db.Exec(query, num, userId)

	if err != nil {
		fmt.Println("✡️  line 268 err", err)
	}
}

func GetUsersAttandance(db *sql.DB, from time.Time, until time.Time) ([]t.UserAttendance, error) {
	var users []t.UserAttendance
	query := `
      SELECT u.id, username, name, emoji, COUNT(user_id) AS mycount FROM yoga.attendance 
      JOIN yoga.user u ON u.id = user_id WHERE date>= $1 and date <= $2 
      GROUP BY u.id, username, name, emoji ORDER BY mycount DESC;`

	rows, err := db.Query(query, from, until)
	defer rows.Close()

	if err != nil {
		return users, err
	}

	for rows.Next() {
		var u t.UserAttendance
		var username sql.NullString

		err := rows.Scan(&u.U.ID, &username, &u.U.Name, &u.U.Emoji, &u.Count)

		if username.Valid {
			u.U.Username = username.String
		}

		if err != nil {
			return users, err
		} else {
			users = append(users, u)
		}
	}

	return users, nil
}

func UpdateLessonDate(db *sql.DB, lessonId int, date string) {
	query := `UPDATE yoga.lesson SET date=$1 WHERE id=$2`

	db.Query(query, date, lessonId)
}

func UpdateLessonTime(db *sql.DB, lessonId int, time string) {
	query := `UPDATE yoga.lesson SET time=$1 WHERE id=$2`

	db.Query(query, time, lessonId)
}

func UpdateLessonDesc(db *sql.DB, lessonId int, description string) {
	query := `UPDATE yoga.lesson SET description=$1 WHERE id=$2`

	db.Query(query, description, lessonId)
}

func UpdateLessonMax(db *sql.DB, lessonId, max int) {
	query := `UPDATE yoga.lesson SET max=$1 WHERE id=$2`

	db.Query(query, max, lessonId)
}

func UpdateUserBio(db *sql.DB, userId int64, userName, fullName string) {
	query := `UPDATE yoga.user SET username=$1, name=$2 WHERE id=$3;`

	_, err := db.Exec(query, userName, fullName, userId)

	if err != nil {
		fmt.Println("✡️  line 348 err", err)
	}
}

func IsLessonSigned(db *sql.DB, lessonId int) bool {
	var id int
	query := "SELECT id FROM yoga.attendance WHERE lesson_id=$1 limit 1;"

	err := db.QueryRow(query, lessonId).Scan(&id)

	if err == sql.ErrNoRows {
		return false
	}

	if err != nil {
		fmt.Println("✡️  line 363 err", err)
		return false
	}

	return true
}

func BlockUsers(db *sql.DB, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	query := "UPDATE yoga.user SET is_blocked = true WHERE id = ANY($1);"

	_, err := db.Exec(query, pq.Int64Array(ids))

	return err
}

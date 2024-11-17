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

func GetLessonWithUsers(db *sql.DB, callBackData string) (*t.LessonWithUsers, error) {
	query := `SELECT u.id, username, name, emoji, l.id, time, date, description, max
		FROM yoga.lesson l LEFT JOIN yoga.registered_users r ON l.id = r.lesson_id
		LEFT JOIN yoga.user u ON u.id = ANY(r.registered) WHERE l.id=$1;`

	data := strings.Split(callBackData, "=")

	rows, err := db.Query(query, data[1])
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	var l t.Lesson
	var users []t.UserDB
	initializedLesson := false

	for rows.Next() {
		var user t.UserDB
		var userId sql.NullInt64
		var username sql.NullString
		var name sql.NullString
		var emoji sql.NullString

		err := rows.Scan(
			&userId, &username, &name, &emoji,
			&l.ID, &l.Time, &l.Date, &l.Description, &l.Max,
		)

		if err != nil {
			return nil, fmt.Errorf("✡️  line 51 err", err)
		}

		if userId.Valid {
			user.ID = userId.Int64
			user.Username = &username.String
			user.Name = name.String
			user.Emoji = emoji.String
		}

		if user.ID != 0 {
			users = append(users, user)
		}

		// Initialize lesson only once
		if !initializedLesson {
			initializedLesson = true
		}
	}

	defer rows.Close()

	return &t.LessonWithUsers{
		Lesson: l,
		Users:  users,
	}, nil
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
	username = EXCLUDED.username, name = EXCLUDED.name;`

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
	WHERE m.ends >= NOW() - INTERVAL '2 months';`

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

func FindUsersByName(db *sql.DB, name string) []t.UserDB {
	var users []t.UserDB
	query := `SELECT * FROM yoga.user WHERE name ILIKE '%' || $1 || '%' OR username ILIKE '%' || $1 || '%';`

	rows, err := db.Query(query, name)

	if err == nil {
		for rows.Next() {
			var u t.UserDB

			err := rows.Scan(&u.ID, &u.Username, &u.Name, &u.Emoji)

			if err != nil {
				fmt.Println("✡️  line 171 err", err)
			}

			users = append(users, u)
		}
	}
	defer rows.Close()

	return users
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

	rows, err := db.Query("SELECT id FROM yoga.user;")
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

func GetUsersAttandance(db *sql.DB, from time.Time, until time.Time) []t.UserAttendance {
	var users []t.UserAttendance
	query := `
      SELECT u.id, username, name, emoji, COUNT(user_id) AS mycount FROM yoga.attendance 
      JOIN yoga.user u ON u.id = user_id WHERE date>= $1 and date <= $2 
      GROUP BY u.id, username, name, emoji ORDER BY mycount DESC;`

	rows, err := db.Query(query, from, until)
	defer rows.Close()

	if err == nil {
		for rows.Next() {
			var u t.UserAttendance

			err := rows.Scan(&u.U.ID, &u.U.Username, &u.U.Name, &u.U.Emoji, &u.Count)

			if err != nil {
				fmt.Println("✡️  line 305 err", err)
			} else {
				users = append(users, u)
			}

		}
	} else {
		fmt.Println("✡️  line 312 err", err)

	}

	return users
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

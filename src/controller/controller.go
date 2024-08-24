package controller

import (
	"bot/src/utils"
	t "bot/src/utils/types"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

func GetAvaliableLessons(db *sql.DB) []t.Lesson {
	var lessons []t.Lesson
	query := "SELECT * FROM yoga.available_lessons ORDER BY date ASC;"

	rows, err := db.Query(query)

	if err == nil {
		for rows.Next() {
			var l t.Lesson

			err := rows.Scan(&l.Id, &l.Date, &l.Time, &l.Description, &l.Max)

			if err != nil {
				fmt.Println("✡️  line 22 err", err)
			}

			lessons = append(lessons, l)
		}
	}
	defer rows.Close()

	return lessons
}

func GetLessonWithUsers(db *sql.DB, callBackData string) []t.LessonWithUsers {
	var lessons []t.LessonWithUsers 
	query := "SELECT u.id, username, name, emoji, l.id, time, date, description, max" + 
	" FROM yoga.lesson l" + 
	" LEFT JOIN yoga.registered_users r ON l.id = r.lesson_id" +
	" LEFT JOIN yoga.user u ON u.id = ANY(r.registered)" +
	" WHERE l.id=$1;"

	data := strings.Split(callBackData, "=")

	rows, err := db.Query(query, data[1])

	if err == nil {
		for rows.Next() {
			var l t.LessonWithUsers

			err := rows.Scan(
				&l.UserId, &l.Username, &l.Name, &l.Emoji,
				&l.LessonId, &l.Time, &l.Date, &l.Description, &l.Max,
				)
			
			if err != nil {
				fmt.Println("✡️  line 51 err", err)
			}

			lessons = append(lessons, l)
		}
	}
	defer rows.Close()

	return lessons
}

func ToggleUserInLesson(db *sql.DB, u t.Update) string {
	text := ""
	data := strings.Split(u.CallbackQuery.Data, "=")
	action := data[0]
	lessonId := data[1]

	switch action {
	case "REGISTER":
		query := "UPDATE yoga.registered_users" +
		" SET registered = array_append(registered, $1)" +
		" WHERE lesson_id=$2 AND NOT ($1=ANY(registered));"

		db.Exec(query, u.CallbackQuery.From.ID, lessonId)

		text = "See you in the session✨"
	case "UNREGISTER":
		query := "UPDATE yoga.registered_users" +
		" SET registered = array_remove(registered, $1)" +
		" WHERE lesson_id=$2 AND NOT ($1=ANY(registered));"

		db.Exec(query, u.CallbackQuery.From.ID, lessonId)

		text = "You are free, fatass...🌚"
	}

	return text
}

func CreateLesson(){ 
	//TODO don't forget to add a row into registered_users
}

func CreateToken(db *sql.DB, tokenType string) string {
	uuid := uuid.New()
	created := time.Now()
	query := "INSERT INTO yoga.token (id, type, created, valid) VALUES ($1,$2,$3,$4);"
	
	_, err := db.Query(query, uuid, tokenType, created, true)

	if err == nil {
		return uuid.String()
	} else {
		return utils.Wrong
	}
}

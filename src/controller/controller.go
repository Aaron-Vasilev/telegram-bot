package controller

import (
	t "bot/src/utils/types"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

func GetAvaliableLessons(db *sql.DB) []t.Lesson {
	var lessons []t.Lesson
	query := "SELECT * FROM yoga.available_lessons ORDER BY date ASC;"

	rows, err := db.Query(query)

	if err == nil {
		for rows.Next() {
			var l t.Lesson

			rows.Scan(&l.ID, &l.Date, &l.Time, &l.Description, &l.Max)

			lessons = append(lessons, l)
		}
	}
	defer rows.Close()

	return lessons
}

type getLessonWithUsersRes struct {
	Data t.LessonWithUsers `json:"get_lesson"`
}

func GetLessonWithUsers(db *sql.DB, s string) t.LessonWithUsers {
	var res getLessonWithUsersRes
	query := "select lesson_id, id, username, name, emoji from yoga.registered_users r join yoga.user u on u.id = any(r.registered) where lesson_id = 21;"
	data := strings.Split(s, "T")
	currTime := time.Now()

	date := fmt.Sprintf("%s-%s", data[0], currTime.Format("2006"))
	time := data[1]

	row, err := db.Query(query, date, time)
	fmt.Println("✡️  line 45 err", err)

	for row.Next() {
		err = row.Scan(&res.Data)
		fmt.Println("✡️  line 48 err", err)
	}

	return res.Data
}

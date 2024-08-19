package controller

import (
	t "bot/src/utils/types"
	"database/sql"
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

// func GetLessonWithUsers(db *sql.DB, ) {
// 	var lesson []t.LessonWithUsers
// 	query := "SELECT * FROM yoga.get_lesson($1, $2);"

// 	rows, err := db.Query(query, )
// }

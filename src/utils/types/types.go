package t

import (
	"time"
)

type CustomError struct {
	Message string
	Error   error
}

type UserDB struct {
  Id        int64  `json:"id"`
  Name      string `json:"name"`
  Username *string `json:"username,omitempty"`
  Emoji     string `json:"emoji"`
}

type Lesson struct {
	ID          int       `json:"id"`
	Date        time.Time `json:"date"`
	Time        time.Time `json:"time"`
	Description string    `json:"description"`
	Max         int       `json:"max"`
}

type LessonWithUsers struct {
	User
	LessonID   int    `json:"lessonId"`
	Registered []User `json:"registered"`
}

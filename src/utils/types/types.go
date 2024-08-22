package t

import (
	"time"
)

type CustomError struct {
	Message string
	Error   error
}

type UserDB struct {
	Id        int64 
	Name      string
	Username *string
	Emoji     string
}
type Lesson struct {
	Id          int      
	Date        time.Time
	Time        time.Time
	Description string   
	Max         int      
}

type LessonWithUsers struct {
	UserId     *int64 
	Name       *string
	Username   *string
	Emoji      *string
	LessonId    int
	Date        time.Time
	Time        time.Time
	Description string   
	Max         int      
}

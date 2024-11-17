package t

import (
	"time"

	"github.com/google/uuid"
)

type CustomError struct {
	Message string
	Error   error
}

type UserDB struct {
	ID       int64
	Name     string
	Username *string
	Emoji    string
}
type Lesson struct {
	ID          int
	Date        time.Time
	Time        time.Time
	Description string
	Max         int
}

type LessonWithUsers struct {
	Users  []UserDB
	Lesson Lesson
}

type Membership struct {
	UserID           int64
	Starts           time.Time
	Ends             time.Time
	Type             int
	LessonsAvailable int
}

type UserMembership struct {
	User             UserDB
	Starts           *time.Time
	Ends             *time.Time
	Type             *int
	LessonsAvailable *int
}

type Token struct {
	ID      uuid.UUID
	Type    int
	Created time.Time
	Valid   bool
}

type RegisterdOnLesson struct {
	IDs      []int64
	LessonId int
	Date     time.Time
}

type UserAttendance struct {
	U     UserDB
	Count int
}

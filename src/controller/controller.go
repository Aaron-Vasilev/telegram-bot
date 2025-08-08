package controller

import (
	"bot/src/db"
	"bot/src/utils"
	t "bot/src/utils/types"
	"context"
	"time"
)

func ToggleUserInLesson(ctx context.Context, userId int64, lessonId int, action string) {
	switch action {
	case "REGISTER":
		db.Query.RegisterUser(ctx, db.RegisterUserParams{
			ArrayAppend: userId,
			LessonID:    lessonId,
		})
	case "UNREGISTER":
		db.Query.UnregisterUser(ctx, db.UnregisterUserParams{
			ArrayRemove: userId,
			LessonID:    lessonId,
		})
	}
}

func AddLesson(ctx context.Context, params db.AddLessonParams) error {
	id, err := db.Query.AddLesson(ctx, params)

	if err != nil {
		return err
	}

	return db.Query.AddRegisterdUsersRow(ctx, db.AddRegisterdUsersRowParams{
		LessonID:   id,
		Registered: make([]int64, 0),
	})
}

func UpdateMembership(
	ctx context.Context,
	userId int64,
	memType int,
) t.Membership {
	var m t.Membership
	token := t.Token{
		Created: time.Now(),
		Type:    memType,
	}

	mem, err := db.Query.GetMembership(ctx, userId)

	if err != nil {
		tmp, _ := time.Parse("2006-01-02", "2023-01-01")
		m.Starts = tmp
		m.Ends = tmp
	} else {
		m.Starts = mem.Starts
		m.Ends = mem.Ends
		m.LessonsAvailable = mem.LessonsAvaliable
		m.Type = mem.Type
	}
	m.UserID = userId

	utils.UpdateMembership(&m, token)

	db.Query.UpdateMembership(ctx, db.UpdateMembershipParams{
		UserID:           userId,
		Type:             memType,
		Starts:           m.Starts,
		Ends:             m.Ends,
		LessonsAvaliable: m.LessonsAvailable,
	})

	return m
}

func IsLessonSigned(c context.Context, lessonId int) bool {
	id, _ := db.Query.GetFirstLessonAttendanceId(c, lessonId)

	if id == 0 {
		return false
	} else {
		return true
	}
}

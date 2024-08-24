package handler

import (
	"bot/src/action"
	"bot/src/bot"
	"bot/src/scene"
	"bot/src/utils"
	t "bot/src/utils/types"
	"context"
	"database/sql"
	"regexp"
)

func handleCallbackQuery(bot *bot.Bot, db *sql.DB, u t.Update) {
	data := u.CallbackQuery.Data
	timetableRe := regexp.MustCompile(`^SHOW_LESSON=\d+$`)
	lessonRe := regexp.MustCompile(`^(REGISTER|UNREGISTER)=\d+$`)

	if timetableRe.MatchString(data) {
		action.SendLesson(bot, db, u)
	} else if lessonRe.MatchString(data) {
		action.RegisterForLesson(bot, db, u)
	}
}

func handleScene(ctx context.Context, bot *bot.Bot, db *sql.DB, u t.Update) context.Context {
	userId, _ := utils.UserIdFromUpdate(u) 

	state, _ := ctx.Value(userId).(scene.SceneState)

	switch  state.Scene {
	case utils.GenerateToken: 
		return scene.GenTokenScene(ctx, bot, db, u)
	default: 
		return ctx
	}
}

func handleAdminCmd(ctx context.Context, bot *bot.Bot, db *sql.DB, u t.Update) context.Context {
	switch u.Message.Text {
	case "ADMIN": 
		action.SendAdminKeyboard(bot, u.Message.From.ID)
	case utils.GenerateToken:
		sceneCtx := context.WithValue(ctx, u.Message.From.ID, scene.SceneState{
			Scene: utils.GenerateToken,
			Stage: 1,
		})

		return scene.GenTokenScene(sceneCtx, bot, db, u)
	}

	return ctx
}

func handleKeyboard(bot *bot.Bot, db *sql.DB, u t.Update) {
	key := u.Message.Text

	switch key {
	case utils.Timetable: 
		action.SendTimetable(bot, db, u)
	case utils.Leaderboard: 
// 		action.SendTimetable(bot, db, u)
	case utils.Profile: 
// 		action.SendTimetable(bot, db, u)
	case utils.Contact: 
// 		action.SendTimetable(bot, db, u)
	}
}

func handleUpdates(ctx context.Context, bot *bot.Bot, db *sql.DB, u t.Update) context.Context {
	if u.CallbackQuery == nil && u.Message == nil {
// 		handleMenu()
		action.SendKeyboard(bot, u.MyChatMember.From.ID, utils.Greeting)

		return ctx
	}

	userId, updateWithCallbackQuery := utils.UserIdFromUpdate(u) 

	_, ok := ctx.Value(userId).(scene.SceneState)

	if ok {
		handleScene(ctx, bot, db, u)
	} else if updateWithCallbackQuery {
		handleCallbackQuery(bot, db, u)
	} else if _, exists := utils.Keyboard[u.Message.Text]; exists {
		handleKeyboard(bot, db, u)
	} else if utils.IsAdmin(userId) {
		return handleAdminCmd(ctx, bot, db, u)
	}

	return ctx
}

func HandleUpdates(c context.Context, bot *bot.Bot, db *sql.DB, updates []t.Update) context.Context {
	for _, update := range updates {
		ctx := handleUpdates(c, bot, db, update)

		bot.Offset = update.UpdateID + 1

		return ctx
	}

	return c
}


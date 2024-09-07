package handler

import (
	"bot/src/action"
	"bot/src/bot"
	"bot/src/controller"
	"bot/src/scene"
	"bot/src/utils"
	t "bot/src/utils/types"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
)

func handleCallbackQuery(ctx *scene.Ctx, bot *bot.Bot, db *sql.DB, u t.Update) {
	data := u.CallbackData()
	timetableRe := regexp.MustCompile(`^SHOW_LESSON=\d+$`)
	lessonRe := regexp.MustCompile(`^(REGISTER|UNREGISTER)=\d+$`)

	if data == utils.ChangeEmoji {
		ctx.Start(u.FromChat().ID, utils.ChangeEmoji)

		scene.ChangeEmoji(ctx, bot, db, u)
	} else if data == utils.HowToFind {
		action.SendHowToFind(bot, db, u)
	} else if data == utils.Timetable {
		action.SendTimetable(bot, db, u)
	} else if timetableRe.MatchString(data) {
		action.SendLesson(bot, db, u)
	} else if lessonRe.MatchString(data) {
		action.RegisterForLesson(bot, db, u)
	}
}

func handleScene(ctx *scene.Ctx, bot *bot.Bot, db *sql.DB, u t.Update) {
	state, _ := ctx.GetValue(u.FromChat().ID)

	switch state.Scene {
	case utils.SignStudents:
		scene.SignStudents(ctx, bot, db, u)
	case utils.ChangeEmoji:
		scene.ChangeEmoji(ctx, bot, db, u)
	case utils.AddLessons:
		scene.AddLessons(ctx, bot, db, u)
	case utils.AssignMembership:
		scene.AssignMembership(ctx, bot, db, u)
	case utils.NotifyAboutLessons:
		scene.NotifyAboutLessons(ctx, bot, db, u)
	case utils.ExtendMemDate:
		scene.ExtendMemEndDate(ctx, bot, db, u)
	}
}

func handleAdminCmd(ctx *scene.Ctx, bot *bot.Bot, db *sql.DB, u t.Update) {
	switch u.Message.Text {
	case "ADMIN":
		action.SendAdminKeyboard(bot, u.Message.From.ID)
	case "USER":
		action.SendKeyboard(bot, u.Message.From.ID, "User Keyboard")
	case utils.SignStudents:
		ctx.SetValue(u.Message.From.ID, scene.SceneState{
			Scene: utils.SignStudents,
			Stage: 1,
		})

		scene.SignStudents(ctx, bot, db, u)
	case utils.AddLessons:
		ctx.SetValue(u.Message.From.ID, scene.SceneState{
			Scene: utils.AddLessons,
			Stage: 1,
		})

		scene.AddLessons(ctx, bot, db, u)
	case utils.AssignMembership:
		ctx.SetValue(u.Message.From.ID, scene.SceneState{
			Scene: utils.AssignMembership,
			Stage: 1,
		})

		scene.AssignMembership(ctx, bot, db, u)
	case utils.NotifyAboutLessons:
		ctx.SetValue(u.Message.From.ID, scene.SceneState{
			Scene: utils.NotifyAboutLessons,
			Stage: 1,
		})

		scene.NotifyAboutLessons(ctx, bot, db, u)
	case utils.ExtendMemDate:
		ctx.SetValue(u.Message.From.ID, scene.SceneState{
			Scene: utils.ExtendMemDate,
			Stage: 1,
		})

		scene.ExtendMemEndDate(ctx, bot, db, u)
	}
}

func handleMenu(bot *bot.Bot, db *sql.DB, u t.Update) {
	if u.FromChat() == nil {
		user := u.MyChatMember.From
		name := strings.Trim(fmt.Sprintf("%s %s", user.FirstName, user.LastName), " ")

		controller.SaveUser(db, user.ID, user.UserName, name)
		action.SendKeyboard(bot, user.ID, utils.GreetingMsg)
	} else if u.Message.Text == "/start" {
		user := u.Message.From
		name := strings.Trim(fmt.Sprintf("%s %s", user.FirstName, user.LastName), " ")

		controller.SaveUser(db, user.ID, user.UserName, name)
		action.SendKeyboard(bot, user.ID, utils.GreetingMsg)
	}
}

func handleKeyboard(bot *bot.Bot, db *sql.DB, u t.Update) {
	key := u.Message.Text

	switch key {
	case utils.Timetable:
		action.SendTimetable(bot, db, u)
	case utils.Leaderboard:
		bot.SendText(u.FromChat().ID, "Work is in progress🛠️")
	case utils.Profile:
		action.SendProfile(bot, db, u.FromChat().ID)
	case utils.Contact:
		action.SendContact(bot, u)
	}
}

func handleUpdates(ctx *scene.Ctx, bot *bot.Bot, db *sql.DB, u t.Update) {
	if u.FromChat() == nil || (u.Message != nil && strings.HasPrefix(u.Message.Text, "/")) {
		handleMenu(bot, db, u)

		return
	}

	userId, updateWithCallbackQuery := utils.UserIdFromUpdate(u)

	_, ok := ctx.GetValue(userId)

	if ok {
		handleScene(ctx, bot, db, u)
	} else if updateWithCallbackQuery {
		handleCallbackQuery(ctx, bot, db, u)
	} else if _, exists := utils.Keyboard[u.Message.Text]; exists {
		handleKeyboard(bot, db, u)
	} else if strings.HasPrefix(u.Message.Text, "/") {
		handleMenu(bot, db, u)
	} else if utils.IsAdmin(userId) {
		handleAdminCmd(ctx, bot, db, u)
	}
}

func HandleUpdates(c *scene.Ctx, bot *bot.Bot, db *sql.DB, updates []t.Update) {
	for _, update := range updates {
		handleUpdates(c, bot, db, update)

		bot.Offset = update.UpdateID + 1
	}
}

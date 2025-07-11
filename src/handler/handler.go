package handler

import (
	"bot/src/action"
	"bot/src/bot"
	"bot/src/controller"
	"bot/src/scene"
	"bot/src/utils"
	t "bot/src/utils/types"
	"database/sql"
	"regexp"
	"slices"
	"strings"
)

func handleCallbackQuery(ctx *scene.Ctx, bot *bot.Bot, db *sql.DB, u t.Update) {
	data := u.CallbackData()
	lessonRe := regexp.MustCompile(`^(REGISTER|UNREGISTER)=\d+$`)
	ifUserComesRe := regexp.MustCompile(`^(IF_USER_COMES)=\d+=(YES|NO)$`)

	if data == utils.ChangeEmoji {
		scene.Start(ctx, bot, db, u, utils.ChangeEmoji)
	} else if data == utils.HowToFind {
		action.SendHowToFind(bot, db, u)
	} else if data == utils.Timetable {
		action.SendTimetable(bot, db, u)
	} else if utils.LessonRegexp().MatchString(data) {
		action.SendLesson(bot, db, u)
	} else if lessonRe.MatchString(data) {
		action.RegisterForLesson(bot, db, u)
	} else if ifUserComesRe.MatchString(data) {
		action.IfUserComesHandler(bot, db, u)
	}
}

func handleScene(ctx *scene.Ctx, bot *bot.Bot, db *sql.DB, u t.Update) {
	state, _ := ctx.GetValue(u.FromChat().ID)

	scene.Map[state.Scene](ctx, bot, db, u)
}

func handleAdminCmd(ctx *scene.Ctx, bot *bot.Bot, db *sql.DB, u t.Update) {
	if u.Message == nil {
		return
	}

	cmd := u.Message.Text
	if _, exist := scene.Map[cmd]; exist {

		scene.Start(ctx, bot, db, u, cmd)
		return
	}
	switch u.Message.Text {
	case "ADMIN":
		action.SendAdminKeyboard(bot, u.Message.From.ID)
	case "USER":
		action.SendKeyboard(bot, u.Message.From.ID, "User Keyboard")
	}
}

func handleMenu(bot *bot.Bot, db *sql.DB, u t.Update) {
	if u.FromChat() == nil || u.Message.Text == "/start" {
		var user t.User

		if u.FromChat() == nil {
			user = u.MyChatMember.From
		} else {
			user = *u.Message.From
		}

		fullName := utils.FullName(user.FirstName, user.LastName)
		controller.SaveUser(db, user.ID, user.UserName, fullName)
		action.SendKeyboard(bot, user.ID, utils.GreetingMsg)
	}
}

func handleKeyboard(bot *bot.Bot, db *sql.DB, u t.Update) {
	key := u.Message.Text

	switch key {
	case utils.Timetable:
		action.SendTimetable(bot, db, u)
	case utils.Leaderboard:
		action.SendLeaderboard(bot, db, u.FromChat().ID)
	case utils.Profile:
		action.SendProfile(bot, db, u.FromChat().ID)
	case utils.Contact:
		action.SendContact(bot, u)
	case utils.Prices:
		action.SendPrices(bot, u)
	case utils.Course:
		action.CourseAction(bot, db, u)
	}
}

func HandleUpdate(ctx *scene.Ctx, bot *bot.Bot, db *sql.DB, u t.Update) {
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
	} else if slices.Contains(utils.Keyboard, u.Message.Text) {
		handleKeyboard(bot, db, u)
	} else if strings.HasPrefix(u.Message.Text, "/") {
		handleMenu(bot, db, u)
	} else if utils.IsAdmin(userId) {
		handleAdminCmd(ctx, bot, db, u)
	}
}

func HandleUpdates(c *scene.Ctx, bot *bot.Bot, db *sql.DB, updates []t.Update) {
	for _, update := range updates {
		HandleUpdate(c, bot, db, update)

		bot.Offset = update.UpdateID + 1
	}
}

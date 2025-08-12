package handler

import (
	"bot/src/action"
	"bot/src/bot"
	"bot/src/db"
	"bot/src/scene"
	"bot/src/utils"
	t "bot/src/utils/types"
	"regexp"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
)

func handleCallbackQuery(bot *bot.Bot, u t.Update) {
	data := u.CallbackData()
	lessonRe := regexp.MustCompile(`^(REGISTER|UNREGISTER)=\d+$`)

	if data == utils.ChangeEmoji {
		scene.Start(bot, u, utils.ChangeEmoji)
	} else if data == utils.HowToFind {
		action.SendHowToFind(bot, u)
	} else if data == utils.Timetable {
		action.SendTimetable(bot, u)
	} else if utils.LessonRegexp().MatchString(data) {
		action.SendLesson(bot, u)
	} else if lessonRe.MatchString(data) {
		action.RegisterForLesson(bot, u)
	}
}

func handleScene(bot *bot.Bot, u t.Update) {
	state, _ := bot.GetCtxValue(u.FromChat().ID)

	scene.Map[state.Scene](bot, u)
}

func handleAdminCmd(bot *bot.Bot, u t.Update) {
	if u.Message == nil {
		return
	}

	cmd := u.Message.Text
	if _, exist := scene.Map[cmd]; exist {

		scene.Start(bot, u, cmd)
		return
	}
	switch u.Message.Text {
	case "ADMIN":
		action.SendAdminKeyboard(bot, u.Message.From.ID)
	case "USER":
		action.SendKeyboard(bot, u.Message.From.ID, "User Keyboard")
	}
}

func handleMenu(bot *bot.Bot, u t.Update) {
	if u.FromChat() == nil || u.Message.Text == "/start" {
		var user t.User

		if u.FromChat() == nil {
			user = u.MyChatMember.From
		} else {
			user = *u.Message.From
		}

		fullName := utils.FullName(user.FirstName, user.LastName)

		params := db.UpsertUserParams{
			ID:   user.ID,
			Name: fullName,
		}

		if user.UserName != "" {
			params.Username = pgtype.Text{
				String: user.UserName,
				Valid:  true,
			}
		}

		db.Query.UpsertUser(bot.Ctx, params)
		action.SendKeyboard(bot, user.ID, utils.GreetingMsg)
	}
}

func handleKeyboard(bot *bot.Bot, u t.Update) {
	key := u.Message.Text

	switch key {
	case utils.Timetable:
		action.SendTimetable(bot, u)
	case utils.Leaderboard:
		action.SendLeaderboard(bot, u.FromChat().ID)
	case utils.Profile:
		action.SendProfile(bot, u.FromChat().ID)
	case utils.Contact:
		action.SendContact(bot, u)
	case utils.Prices:
		action.SendPrices(bot, u)
	case utils.Course:
		action.CourseAction(bot, u)
	}
}

func HandleUpdate(bot *bot.Bot, u t.Update) {
	if u.FromChat() == nil {
		if u.PollAnswer != nil {
			handlePool(bot, u)

			return
		} else if u.Message == nil || strings.HasPrefix(u.Message.Text, "/") {
			handleMenu(bot, u)

			return 
		}
	}

	userId, updateWithCallbackQuery := utils.UserIdFromUpdate(u)

	_, ok := bot.GetCtxValue(userId)

	if ok {
		handleScene(bot, u)
	} else if updateWithCallbackQuery {
		handleCallbackQuery(bot, u)
	} else if slices.Contains(utils.Keyboard, u.Message.Text) {
		handleKeyboard(bot, u)
	} else if strings.HasPrefix(u.Message.Text, "/") {
		handleMenu(bot, u)
	} else if utils.IsAdmin(userId) {
		handleAdminCmd(bot, u)
	}
}

func HandleUpdates(bot *bot.Bot, updates []t.Update) {
	for _, update := range updates {
		HandleUpdate(bot, update)

		bot.Offset = update.UpdateID + 1
	}
}

func handlePool(bot *bot.Bot, u t.Update) {
	action.IfUserComesHandler(bot, u.PollAnswer)
}

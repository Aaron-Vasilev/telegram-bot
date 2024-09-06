package action

import (
	"bot/src/bot"
	"bot/src/controller"
	"bot/src/utils"
	t "bot/src/utils/types"
	"database/sql"
)

func SendTimetable(bot *bot.Bot, db *sql.DB, upd t.Update) {
	lessons := controller.GetAvaliableLessons(db)

	msg := utils.GenerateTimetable(lessons, false)
	msg.ChatId = upd.FromChat().ID

	bot.SendMessage(msg)
}

func SendContact(bot *bot.Bot, u t.Update) {
	bot.SendMessage(t.Message{
		ChatId:    u.FromChat().ID,
		Text:      utils.ContactMsg,
		ParseMode: "html",
		ReplyMarkup: t.InlineKeyboardMarkup{
			InlineKeyboard: [][]t.InlineKeyboardButton{
				{
					{
						Text:         utils.HowToFind,
						CallbackData: utils.HowToFind,
					},
				},
			},
		},
	})
}

func SendProfile(bot *bot.Bot, db *sql.DB, chatId int64) {
	userWithMem := controller.GetUserWithMembership(db, chatId)

	buttons := [][]t.InlineKeyboardButton{
		{
			{
				Text:         utils.ChangeEmoji,
				CallbackData: utils.ChangeEmoji,
			},
		},
	}

	msg := t.Message{
		ChatId:    chatId,
		Text:      utils.ProfileText(userWithMem),
		ParseMode: "html",
		ReplyMarkup: t.InlineKeyboardMarkup{
			InlineKeyboard: buttons,
		},
	}

	bot.SendMessage(msg)
}

func SendAdminKeyboard(bot *bot.Bot, chatId int64) {
	replyKeyboard := t.ReplyKeyboardMarkup{
		Keyboard: [][]t.KeyboardButton{
			{
				{
					Text: utils.AdminKeyboard[utils.SignStudents],
				},
			},
			{
				{
					Text: utils.AdminKeyboard[utils.AddLessons],
				},
			},
			{
				{
					Text: utils.AdminKeyboard[utils.AssignMembership],
				},
			},
			{
				{
					Text: utils.AdminKeyboard[utils.NotifyAboutLessons],
				},
			},
		},
		ResizeKeyboard: true,
	}

	msg := t.Message{
		Text:        "Switch to admin mode🃏",
		ChatId:      chatId,
		ReplyMarkup: &replyKeyboard,
	}

	bot.SendMessage(msg)
}

func SendKeyboard(bot *bot.Bot, chatId int64, text string) {
	replyKeyboard := t.ReplyKeyboardMarkup{
		Keyboard: [][]t.KeyboardButton{
			{
				{
					Text: utils.Keyboard[utils.Timetable],
				},
				{
					Text: utils.Keyboard[utils.Leaderboard],
				},
			},
			{
				{
					Text: utils.Keyboard[utils.Profile],
				},
				{
					Text: utils.Keyboard[utils.Contact],
				},
			},
		},
		ResizeKeyboard: true,
	}

	msg := t.Message{
		Text:        text,
		ChatId:      chatId,
		ReplyMarkup: &replyKeyboard,
	}

	bot.SendMessage(msg)
}

func SendLesson(bot *bot.Bot, db *sql.DB, u t.Update) {
	lesson := controller.GetLessonWithUsers(db, u.CallbackQuery.Data)

	msg := utils.GenerateLessonMessage(lesson, u.FromChat().ID)

	bot.SendMessage(msg)
}

func RegisterForLesson(bot *bot.Bot, db *sql.DB, u t.Update) {
	text := controller.ToggleUserInLesson(db, u)
	bot.SendText(u.FromChat().ID, text)
}

func SendHowToFind(bot *bot.Bot, db *sql.DB, u t.Update) {
	bot.SendLocation(u.FromChat().ID, 32.05382162148281, 34.75493749973202)
	bot.SendPhotoById(u.FromChat().ID, "AgACAgIAAxkBAAIVsWbZqoIj1U0WQMX97pezh8NPrvS1AAI03zEb_QAB0EqIuOgvJ2h8SQEAAwIAA3MAAzYE")
	bot.SendPhotoById(u.FromChat().ID, "AgACAgIAAxkBAAIVwWbZ5AiqC497NDhWORiJd5oLx6oqAALZ4DEb_QAB0Eojpa9wdlTtSQEAAwIAA3kAAzYE")
}

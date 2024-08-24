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

func SendAdminKeyboard(bot *bot.Bot, chatId int64) {
	replyKeyboard := t.ReplyKeyboardMarkup{
		Keyboard: [][]t.KeyboardButton{ 
			{
				{
					Text: utils.AdminKeyboard[utils.GenerateToken],
				},
			},
		},
		ResizeKeyboard: true,
	}

	msg := t.Message {
		Text: "Switch to admin mode🃏",
		ChatId: chatId,
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

	msg := t.Message {
		Text: text,
		ChatId: chatId,
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
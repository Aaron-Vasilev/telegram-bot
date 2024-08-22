package commands

import (
	"bot/src/bot"
	"bot/src/controller"
	"bot/src/utils"
	t "bot/src/utils/types"
	"database/sql"
	"regexp"
)

func sendTimetable(bot *bot.Bot, db *sql.DB, upd t.Update) {
	lessons := controller.GetAvaliableLessons(db)

	msg := utils.GenerateTimetable(lessons, false)
	msg.ChatId = upd.FromChat().ID

	bot.SendMessage(msg)
}

func sendKeyboard(bot *bot.Bot, chatId int64, text string) {
	replyKeyboard := t.ReplyKeyboardMarkup{
		Keyboard: [][]t.KeyboardButton{ 
			{
				{
					Text: utils.Keyboard["Timetable 🗓"],
				},
				{
					Text: utils.Keyboard["Leaderboard 🏆"],
				},
			},
			{
				{
					Text: utils.Keyboard["Profile 🧘"],
				},
				{
					Text: utils.Keyboard["Contact 💌"],
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

func sendLesson(bot *bot.Bot, db *sql.DB, u t.Update) {
	lesson := controller.GetLessonWithUsers(db, u.CallbackQuery.Data)

	msg := utils.GenerateLessonMessage(lesson, u.FromChat().ID)

	bot.SendMessage(msg)
}

func registerForLesson(bot *bot.Bot, db *sql.DB, u t.Update) {
	text := controller.ToggleUserInLesson(db, u)
	bot.SendText(u.FromChat().ID, text)
}

func handleCallbackQuery(bot *bot.Bot, db *sql.DB, u t.Update) {
	data := u.CallbackQuery.Data
	timetableRe := regexp.MustCompile(`^SHOW_LESSON=\d+$`)
	lessonRe := regexp.MustCompile(`^(REGISTER|UNREGISTER)=\d+$`)

	if timetableRe.MatchString(data) {
		sendLesson(bot, db, u)
	} else if lessonRe.MatchString(data) {
		registerForLesson(bot, db, u)
	}

}

func handleKeyboard(bot *bot.Bot, db *sql.DB, u t.Update) {
	key := u.Message.Text

	switch key {
	case utils.Timetable: 
		sendTimetable(bot, db, u)
	}
}

func handleMessage(bot *bot.Bot, db *sql.DB, u t.Update) {

	if u.CallbackQuery != nil {
		handleCallbackQuery(bot, db, u)
	} else if _, exists := utils.Keyboard[u.Message.Text]; exists {
		handleKeyboard(bot, db, u)
	}
}

func HandleUpdates(bot *bot.Bot, db *sql.DB, updates []t.Update) {
	for _, update := range updates {
		handleMessage(bot, db, update)

		bot.Offset = update.UpdateID + 1
	}
}

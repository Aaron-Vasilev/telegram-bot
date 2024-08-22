package commands

import (
	"bot/src/bot"
	"bot/src/controller"
	"bot/src/utils"
	t "bot/src/utils/types"
	"database/sql"
	"fmt"
	"regexp"
)

func sendTimetable(bot *bot.Bot, db *sql.DB, upd t.Update) {
	lessons := controller.GetAvaliableLessons(db)

	msg := utils.GenerateTimetable(lessons, false)
	msg.ChatId = upd.FromChat().ID

	bot.SendMessage(msg)
}

func sendKeyboard(bot *bot.Bot, text string) {
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
		ChatId: 362575139,
		ReplyMarkup: &replyKeyboard,
	}

	bot.SendMessage(msg)
}

func sendLesson(bot *bot.Bot, db *sql.DB, u t.Update) {
	lesson := controller.GetLessonWithUsers(db, u.CallbackQuery.Data)
	fmt.Println("✡️  line 54 lesson", lesson)

}

func handleCallbackQuery(bot *bot.Bot, db *sql.DB, u t.Update) {
	data := u.CallbackQuery.Data
	timetableRe := regexp.MustCompile(`^\d{2}-\d{2}T\d{2}:\d{2}$`)

	if timetableRe.MatchString(data) {
		sendLesson(bot, db, u)
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

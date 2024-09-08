package action

import (
	"bot/src/bot"
	"bot/src/controller"
	"bot/src/utils"
	t "bot/src/utils/types"
	"database/sql"
	"fmt"
	"time"
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
			{
				{
					Text: utils.AdminKeyboard[utils.NotifyAll],
				},
			},
			{
				{
					Text: utils.AdminKeyboard[utils.ExtendMemDate],
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

func NotifyAboutSubscriptionEnds(bot *bot.Bot, db *sql.DB) {
	today := time.Now()
	usersMem := controller.GetAllUsersWithMemLatest(db)

	for _, mem := range usersMem {
		text := "My cherry varenichek🥟🍒\n"

		if *mem.Type == utils.NoLimit {
			text += fmt.Sprintf("Kindly reminder, your membership ends <b>%s</b>", mem.Ends.Format("2006-01-02"))
		} else if *mem.LessonsAvailable <= 0 || mem.Ends.Before(today) {
			text += fmt.Sprintf("When you come to my lesson next time, <b>remember to renew your membership</b>😚")
		} else {
			text += fmt.Sprintf(
				"Your membership ends <b>%s</b> and you still have <b>%d</b> lessons🥳\nDon't forget to use them all🧞‍♀️",
				mem.Ends.Format("2006-01-02"),
				*mem.LessonsAvailable,
			)
		}
		text += "\n" + utils.SeeYouMsg

		bot.SendHTML(mem.User.ID, text)
	}
}

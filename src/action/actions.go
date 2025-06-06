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

	msg := utils.GenerateTimetableMsg(lessons, false)
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

	bot.SendMessage(t.Message{
		ChatId:    u.FromChat().ID,
		Photo: "https://bot-telega.s3.il-central-1.amazonaws.com/door.jpg",
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

func SendLeaderboard(bot *bot.Bot, db *sql.DB, chatId int64) {
	now := time.Now()

	firstDay := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	nextMonth := now.Month() + 1
	nextMonthYear := now.Year()
	if nextMonth > 12 {
		nextMonth = 1
		nextMonthYear++
	}
	firstDayNextMonth := time.Date(nextMonthYear, nextMonth, 1, 0, 0, 0, 0, now.Location())
	lastDay := firstDayNextMonth.AddDate(0, 0, -1)

	usersWithCount := controller.GetUsersAttandance(db, firstDay, lastDay)

	bot.SendHTML(chatId, utils.LeaderboardText(usersWithCount, chatId))
}

func SendAdminKeyboard(bot *bot.Bot, chatId int64) {
	var keyboard [][]t.KeyboardButton
	btns := make([]t.KeyboardButton, 2)

	for i, key := range utils.AdminKeyboard {
		position := i % 2

		btns[position] = t.KeyboardButton{
			Text: key,
		}

		if position == 1 {
			keyboard = append(keyboard, btns)
			btns = make([]t.KeyboardButton, 2)
		}

	}

	if btns[1].Text == "" {
		keyboard = append(keyboard, btns)
	}

	replyKeyboard := t.ReplyKeyboardMarkup{
		Keyboard:       keyboard,
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
	lessonWithUsers, err := controller.GetLessonWithUsers(db, u.CallbackQuery.Data)
	chat := u.FromChat()

	if err != nil {
		bot.Error(fmt.Sprintf("Send lesson error: %s. Data: %s", err.Error(), u.CallbackQuery.Data))
	}

	for _, user := range lessonWithUsers.Users {
		if user.ID == chat.ID {
			fullName := utils.FullName(chat.FirstName, chat.LastName) 

			if (user.Username.Valid && user.Username.String != chat.UserName) || fullName != user.Name  {
				controller.UpdateUserBio(db, chat.ID, chat.UserName, fullName)
			}
		}
	}

	msg := utils.GenerateLessonMessage(lessonWithUsers, u.FromChat().ID)

	bot.SendMessage(msg)
}

func RegisterForLesson(bot *bot.Bot, db *sql.DB, u t.Update) {
	text := controller.ToggleUserInLesson(db, u)
	bot.SendText(u.FromChat().ID, text)
}

func SendHowToFind(bot *bot.Bot, db *sql.DB, u t.Update) {
	bot.SendLocation(u.FromChat().ID, 32.049336, 34.752160)

	media := []t.InputMediaPhoto{
		{
			BaseInputMedia: t.BaseInputMedia{
				Type: "photo",
				Media: "https://bot-telega.s3.il-central-1.amazonaws.com/entrence.jpg",
				Caption: "Пожалуйста не смотрите в окна🪟❌👀 к нашим соседям, они оченьстесняются🫣",
			},
		},
		{
			BaseInputMedia: t.BaseInputMedia{
				Type: "photo",
				Media: "https://bot-telega.s3.il-central-1.amazonaws.com/door.jpg",
			},
		},
	}
	
	bot.SendMediaGroup(u.FromChat().ID, media)
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

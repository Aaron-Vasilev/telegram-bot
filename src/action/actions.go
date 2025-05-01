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
	chat := u.FromChat()
	lessonWithUsers, err := controller.GetLessonWithUsers(db, u.CallbackQuery.Data)

	if err != nil {
		bot.Error(err.Error())
		bot.SendText(chat.ID, utils.WrongMsg)

		return
	}

	for _, user := range lessonWithUsers.Users {
		if user.ID == chat.ID {
			fullName := utils.FullName(chat.FirstName, chat.LastName)

			if user.Username != &chat.UserName || fullName != user.Name {
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
	bot.SendLocation(u.FromChat().ID, 32.049304, 34.752149)
	bot.SendPhotoById(u.FromChat().ID, "https://bot-telega.s3.il-central-1.amazonaws.com/entrence.jpg?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Content-Sha256=UNSIGNED-PAYLOAD&X-Amz-Credential=ASIAYOKUBS744FVJ6CCX%2F20250501%2Fil-central-1%2Fs3%2Faws4_request&X-Amz-Date=20250501T131507Z&X-Amz-Expires=300&X-Amz-Security-Token=IQoJb3JpZ2luX2VjEIL%2F%2F%2F%2F%2F%2F%2F%2F%2F%2FwEaDGlsLWNlbnRyYWwtMSJGMEQCIBrt4dpYh6CdXBlm9fxzT%2BDm1Dh0sGQ33Z7U7CdqRMi1AiAYgQUmW2wIRjHBE9ZFXvSc13ATiv0Nehwcnai8mOU0jCrjAgi%2B%2F%2F%2F%2F%2F%2F%2F%2F%2F%2F8BEAEaDDU4MDUzMzcyMTA4MSIMx%2Bu3TUJvnQMiGCGKKrcCtw4VbrDohQG5mWMDSeXu3bC2kd4cLDdw9J86JRbM41s0OnkOQR%2FAu09hwq9pZ%2F8xDzvVu8qGuCADVjXb3OwxnaA3kLpxA%2FIPQ%2BtEFNY7fNPjKxM72UR0nswTcazLIBzqVZtca64b2WkjBvF2GsQobU47ojwTFn90yJeedZcyRbQNbhBcSYUTq4HW82GqOjOqZluvE7hRVWcgKuy3832djZOmA6M9eBFb8nf%2FjxCmj97NXTfDAdRYFMJoz2p9JuiERwXqvM7c1kzPXnuwY32tbvNPiA%2BML%2BtmZwa1GjsXFq8YRD6DUEuhYKcfq7TRiQ28SwIflb38F4r4q%2FntNFQTxSMCs1L0rOgydz5QwB%2BUrgCG0iG2fXGKAwhZ40MqXN5R%2FMn7DpP%2FsHJ66ua6aYZbtLWqmE90SYIw0uXNwAY6rgK1G3n9NvOJod8VUER%2FhMgxQX9QukluVFvDk%2Bj6BlDUe%2B5LaTEOkpYnLhglOBcRxO4DVWUG%2BLTogCP3OusAjwxz8qNBfKwc45W%2BvfddlRppT%2Bs9v3F9vFTXJKMFm4Ryn9%2Bldj%2BPqoyayx%2FuLEbBEYjxb0GoywxkJupOaO6ri9jpERjLh52c8m3ucuY7rJDPKJ9HQ1u8VWuVuXV4kR9P8aTjGqLlt2%2Bu9sV4AmzjxKAp4nYpkXjsWQXsCFozue01DZpnWn16SvJaebiMDVfjge6JK1CKZja2c7zhMq7oxguNjmi6rHUegEAZcSBsvhQry1RbupR47fl77u1zo6WUSQHaQexzJFBCrw1khItL%2Fm4wV3NJ85bFsSMcJaBPgFS517ZagMfXxxyErYSpmcdruw%3D%3D&X-Amz-Signature=9807c8334d84ace9281ce68626206fea4b6826a48d5a71fb2ee18590ad1e31ad&X-Amz-SignedHeaders=host&response-content-disposition=inline", "")
	bot.SendPhotoById(u.FromChat().ID, "https://bot-telega.s3.il-central-1.amazonaws.com/door.jpg?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Content-Sha256=UNSIGNED-PAYLOAD&X-Amz-Credential=ASIAYOKUBS74ZA4F3G27%2F20250501%2Fil-central-1%2Fs3%2Faws4_request&X-Amz-Date=20250501T132120Z&X-Amz-Expires=300&X-Amz-Security-Token=IQoJb3JpZ2luX2VjEIL%2F%2F%2F%2F%2F%2F%2F%2F%2F%2FwEaDGlsLWNlbnRyYWwtMSJHMEUCIQC7WbgODmTKv5Hwl3rfTIsZHMiriLAslSi%2F5voq9faRTgIgUKgy5GV5wiIrArwrBS6YEXcVu%2FWRWmLyw9bfTMloIH8q4wIIvv%2F%2F%2F%2F%2F%2F%2F%2F%2F%2FARABGgw1ODA1MzM3MjEwODEiDMxhmHWSH708lBqTsyq3AvyErrr9R1EpieG3bUSBkp3O%2FAiflbXYv3FQmTIunW9D9WBSAMbfDMEX4frjO1dtX8gpJVn2tcrGrhP6NvbXNtc8oiDfzt%2FnjxcXuZrQqA11mnqjpRCCjFBSRCPXQxFGRpgGdbUJO8QxjnBtRQ%2BYE3ZTGkO27AjGU8SkB5FnODEwUIGuqTglsn%2FU8ftWg85IEhfPUbw9w38EikRsI4a%2F5PPP7WNpkwqQ9Xps5hnEWGC2x4FvVoXzSrzfnmfGKHLtGmDBcWpuepqY004F066BIwhMPInL1R1kCHOw23fVEO9Q43crU36a6eZXQrHaJTXZS6xP5picqaq3NZSc5JA1n13wUDNpR418p9jS1hM69hqLcCRmQNsKPs8d6EDTBtUZyAUVc97i0DwIrZX8SyJ7t371z%2BCJeaB3MNLlzcAGOq0Cs3iivzH69sS119v83azitvf0JibZSPLoXTsCzhZ%2FhLW71GJc%2BowGtNAHWmcU7z2BYs16a2BRUSCcKGDjZfWpXFgwdH1OCeF5iyXlv%2BG5GeQngkgbPrgpTjYugn9PL7TSneno7QaPIRuaGI3GfBop1ZZpTNFU%2FGebj3mk34UZ8sJwE1PxTEFXvP%2BDY2bQEIVSHYGqcXitKsrpmug1f5HRIe%2B%2FSN%2Flywg3IlGJe74p4lM7ASwvS2RN5FL1GJVkURcPQK3fSoYYUw3m0r5wYY20f4UgXH9GAAFB5Sy%2FmRR2qNxkK%2BS1TMP7FtA%2Fbkz6b8zz4JAXtfsz5tzIN3YhT3JCmKn1w4eeaptA%2BbHu%2BSI64qwYxsTOWZi00gBR1Sag%2BlYOFKPe3h4RZY5pxWamcQ%3D%3D&X-Amz-Signature=e4f9288fd60cd266c362b161e82bc59a6f734f1c56f1d31363413adb9817a08d&X-Amz-SignedHeaders=host&response-content-disposition=inline", "Please do not stare at our neighbors 🪟👀\nThey are really shy🫣")
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

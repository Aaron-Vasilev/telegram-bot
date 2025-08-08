package action

import (
	"bot/src/bot"
	"bot/src/controller"
	"bot/src/db"
	"bot/src/utils"
	t "bot/src/utils/types"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func SendTimetable(bot *bot.Bot, u t.Update) {
	lessons, err := db.Query.GetAvailableLessons(bot.Ctx)

	if err != nil {
		bot.SendText(u.FromChat().ID, utils.WrongMsg)
		bot.Error("Send timetable err: " + err.Error())
		return
	}
	msg := utils.GenerateTimetableMsg(lessons, false)
	msg.ChatId = u.FromChat().ID

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
		ChatId: u.FromChat().ID,
		Photo:  "https://bot-telega.s3.il-central-1.amazonaws.com/door.jpg",
	})
}

func SendPrices(bot *bot.Bot, u t.Update) {
	bot.SendMessage(t.Message{
		ChatId:    u.FromChat().ID,
		Text:      utils.PricesMsg,
		ParseMode: "html",
	})
}

func SendProfile(bot *bot.Bot, chatId int64) {
	userWithMem, err := db.Query.GetUserWithMembership(bot.Ctx, chatId)

	if err != nil {
		bot.Error("Get user with memb error: " + err.Error())
		bot.SendText(chatId, utils.WrongMsg)

		return
	}

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

func SendLeaderboard(bot *bot.Bot, chatId int64) {
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

	usersWithCount, err := db.Query.GetUsersAttandance(bot.Ctx, db.GetUsersAttandanceParams{
		FromDate: firstDay,
		ToDate:   lastDay,
	})

	if err != nil {
		bot.Error("send leaderboard error: " + err.Error())
	}

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
	var keyboard [][]t.KeyboardButton
	var pair []t.KeyboardButton

	for i := range utils.Keyboard {
		if len(pair) == 2 {
			keyboard = append(keyboard, slices.Clone(pair))
			pair = pair[:0]
		}

		pair = append(pair, t.KeyboardButton{
			Text: utils.Keyboard[i],
		})
	}
	keyboard = append(keyboard, pair)
	replyKeyboard := t.ReplyKeyboardMarkup{
		Keyboard:       keyboard,
		ResizeKeyboard: true,
	}

	msg := t.Message{
		Text:        text,
		ChatId:      chatId,
		ReplyMarkup: &replyKeyboard,
	}

	bot.SendMessage(msg)
}

func SendLesson(bot *bot.Bot, u t.Update) {
	lessonId, err := strconv.Atoi(strings.Split(u.CallbackQuery.Data, "=")[1])

	if err != nil {
		bot.Error("send lesson err: " + err.Error())
	}

	lessonWithUsers, err := db.Query.GetLessonWithUsers(bot.Ctx, lessonId)
	chat := u.FromChat()

	if err != nil {
		bot.Error(fmt.Sprintf("Send lesson error: %s. Data: %s", err.Error(), u.CallbackQuery.Data))
		bot.SendText(chat.ID, utils.WrongMsg)
		return
	}

	for _, l := range lessonWithUsers {
		if l.UserID.Valid && l.UserID.Int64 == chat.ID {
			fullName := utils.FullName(chat.FirstName, chat.LastName)

			if (
				(l.Username.Valid && l.Username.String != chat.UserName) || 
				(l.Name.Valid && fullName != l.Name.String)) {
				db.Query.UpdateUserBio(bot.Ctx, db.UpdateUserBioParams{
					ID:   chat.ID,
					Name: fullName,
					Username: pgtype.Text{
						String: chat.UserName,
						Valid:  true,
					},
				})
			}
		}
	}

	msg := utils.GenerateLessonMessage(lessonWithUsers, u.FromChat().ID)

	bot.SendMessage(msg)
}

func RegisterForLesson(bot *bot.Bot, u t.Update) {
	text := ""
	data := strings.Split(u.CallbackQuery.Data, "=")
	action := data[0]
	lessonId, err := strconv.Atoi(data[1])

	if err != nil {
		bot.Error(fmt.Sprintf("Wrong lesson id for lesson: %s", data[1]))
		bot.SendText(u.FromChat().ID, utils.WrongMsg)
		return
	}

	switch action {
	case utils.REGISTER:
		text = utils.SeeYouMsg
	case utils.UNREGISTER:
		text = utils.YouAreFree
	default:
		bot.Error(fmt.Sprintf("Error this action doesn't exists: %s", action))
		bot.SendText(u.FromChat().ID, utils.WrongMsg)
		return
	}

	controller.ToggleUserInLesson(bot.Ctx, u.FromChat().ID, lessonId, action)

	bot.SendText(u.FromChat().ID, text)
}

func SendHowToFind(bot *bot.Bot, u t.Update) {
	bot.SendLocation(u.FromChat().ID, 32.049336, 34.752160)

	media := []t.InputMediaPhoto{
		{
			BaseInputMedia: t.BaseInputMedia{
				Type:    "photo",
				Media:   "https://bot-telega.s3.il-central-1.amazonaws.com/entrence.jpg",
				Caption: "Пожалуйста не смотрите в окна🪟❌👀 к нашим соседям, они оченьстесняются🫣",
			},
		},
		{
			BaseInputMedia: t.BaseInputMedia{
				Type:  "photo",
				Media: "https://bot-telega.s3.il-central-1.amazonaws.com/door.jpg",
			},
		},
	}

	bot.SendMediaGroup(u.FromChat().ID, media)
}

func NotifyAboutSubscriptionEnds(bot *bot.Bot) {
	today := time.Now()
	usersMem, err := db.Query.GetAllUsersWithMemLatest(bot.Ctx)

	if err != nil {
		bot.Error("Get all users with mem latest error:" + err.Error())
		return
	}

	for _, mem := range usersMem {
		text := "My cherry varenichek🥟🍒\n"

		if mem.Type == utils.NoLimit {
			text += fmt.Sprintf("Kindly reminder, your membership ends <b>%s</b>", mem.Ends.Format("2006-01-02"))
		} else if mem.LessonsAvaliable <= 0 || mem.Ends.Before(today) {
			text += fmt.Sprintf("When you come to my lesson next time, <b>remember to renew your membership</b>😚")
		} else {
			text += fmt.Sprintf(
				"Your membership ends <b>%s</b> and you still have <b>%d</b> lessons🥳\nDon't forget to use them all🧞‍♀️",
				mem.Ends.Format("2006-01-02"),
				mem.LessonsAvaliable,
			)
		}
		text += "\n" + utils.SeeYouMsg

		bot.SendHTML(mem.ID, text)
	}
}

func NotifyAboutTommorowLesson(bot *bot.Bot) {
	tomorrow := time.Now().AddDate(0, 0, 1)

	lessons, err := db.Query.GetLessonsByDate(bot.Ctx, tomorrow)

	if err != nil {
		bot.Error("Error in tomorrow's lesson notification: " + err.Error())
		return
	}

	ViolettaId, err := utils.ViolettaId()

	if err != nil {
		bot.Error("No Violetta's id in .env" + err.Error())
		return
	}

	for _, lesson := range lessons {
		res, _ := bot.SendPool(t.PollMessage{
			ChatId: ViolettaId,
			Poll:   utils.IfUserComesPoll(lesson),
		})

		registeredForLessons, err := db.Query.GetRegisteredUsers(bot.Ctx, lesson.ID)

		if err != nil {
			bot.Error(fmt.Sprintf("Error in lesson: %d\nerror: %s", lesson.ID, err.Error()))
			break
		}

		for _, lessons := range registeredForLessons {
			for _, id := range lessons.Registered {
				bot.Forward(id, res.Chat.ID, res.MessageID)
			}
		}
	}
}

// TODO
func IfUserComesHandler(bot *bot.Bot, u *t.PollAnswer) {

	// 	text := ""
	// 	lessonId, err := strconv.Atoi(data[1])

	// 	if err != nil {
	// 		bot.Error(fmt.Sprintf("Wrong lesson id for lesson: %s", data[1]))
	// 		bot.SendText(u.User.ID, utils.WrongMsg)
	// 		return
	// 	}

	// 	switch response {
	// 	case utils.YES:
	// 		text = "You are the best🏆"
	// 	case utils.NO:
	// 		text = utils.YouAreFree
	// 		controller.ToggleUserInLesson(bot.Ctx, u.User.ID, lessonId, utils.UNREGISTER)
	// 	default:
	// 		bot.Error(fmt.Sprintf("Error the response is nor YES or NO: %s", response))
	// 		bot.SendText(u.User.ID, utils.WrongMsg)
	// 		return
	// 	}

	// bot.SendText(u.User.ID, text)
}

func CourseAction(bot *bot.Bot, u t.Update) {
	bot.SendText(u.FromChat().ID, "Comming 🔜")
	return
	// 	hasAccess, err := db.Query.CheckIfUserHasCourseAcces(bot.Ctx, u.FromChat().ID)

	// 	if err != nil {
	// 		bot.Error("CourseAction:"+err.Error())
	// 		bot.SendText(u.FromChat().ID, utils.WrongMsg)
	// 		return
	// 	}

	// 	if !hasAccess {

	// 		replyKeyboard := t.ReplyKeyboardMarkup{
	// 			Keyboard: [][]t.KeyboardButton{
	// 				{
	// 					{
	// 						Text: "TEST",
	// 					},
	// 				},
	// 			},
	// 			ResizeKeyboard: true,
	// 		}

	// 		msg := t.Message{
	// 			Text:        "SUP",
	// 			ChatId:      u.FromChat().ID,
	// 			ReplyMarkup: &replyKeyboard,
	// 		}

	// 		bot.SendMessage(msg)

	// 	} else {

	// }
}

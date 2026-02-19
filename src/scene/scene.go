package scene

import (
	"bot/src/action"
	"bot/src/bot"
	"bot/src/controller"
	"bot/src/db"
	"bot/src/utils"
	t "bot/src/utils/types"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

type sceneCallback = func(bot *bot.Bot, u t.Update)

var Map = map[string]sceneCallback{
	utils.SignStudents:       SignStudents,
	utils.ChangeEmoji:        ChangeEmoji,
	utils.AddLessons:         AddLessons,
	utils.AssignMembership:   AssignMembership,
	utils.NotifyAboutLessons: NotifyAboutLessons,
	utils.ForwardAll:         ForwardAll,
	utils.FreezeMembership:   FreezeMembership,
	utils.EditLesson:         EditLesson,
}

func Start(b *bot.Bot, u t.Update, scene string) {
	b.SetCtxValue(u.FromChat().ID, bot.SceneState{
		Scene: scene,
		Stage: 1,
	})

	Map[scene](b, u)
}

type signStudentsData struct {
	Data  db.GetRegisteredOnLessonRow
	Index int
}

func SignStudents(bot *bot.Bot, u t.Update) {
	userId, _ := utils.UserIdFromUpdate(u)
	state, ok := bot.GetCtxValue(userId)

	if !ok {
		bot.Error(fmt.Sprintf("No scene for the user: %d", userId))
		bot.EndCtx(userId)
	}

	switch state.Stage {
	case 1:
		lessons, err := db.Query.GetLatestLessons(bot.Ctx, 12)

		if err != nil {
			bot.SendText(userId, utils.WrongMsg)
			bot.EndCtx(userId)
			bot.Error("get avaliable lessons error: " + err.Error())
		}

		msg := utils.GenerateTimetableMsg(lessons, true)
		msg.ChatId = userId
		msg.Text = "Send back the <b>ID</b> of the lesson you want to sign students for"
		msg.ParseMode = "html"
		bot.SendMessage(msg)
	case 2:
		if u.Message == nil {
			bot.SendHTML(userId, "You must send an <b>ID</b>")
			bot.EndCtx(userId)
			return
		}
		lessonId, err := strconv.Atoi(u.Message.Text)

		if err != nil {
			bot.SendText(userId, utils.NotANumberMsg)
			bot.EndCtx(userId)
			return
		}

		wasLessonSigned := controller.IsLessonSigned(bot.Ctx, lessonId)

		if wasLessonSigned {
			bot.SendText(userId, "This lesson was signed⚠️")
		}

		registered, err := db.Query.GetRegisteredOnLesson(bot.Ctx, lessonId)

		if err != nil {
			bot.Error("Error getRegisteredOnLesson: " + err.Error())
			bot.EndCtx(userId)
			return
		}

		if len(registered.Registered) == 0 {
			bot.SendText(userId, "There are no users on this lesson")
			bot.EndCtx(userId)
			return
		}
		state.Data = signStudentsData{
			Data:  registered,
			Index: 0,
		}

		bot.SetCtxValue(userId, state)

		userWithMem, err := db.Query.GetUserWithMembership(bot.Ctx, registered.Registered[0])

		if err != nil {
			bot.SendText(userId, "The ID is not correct")
			bot.EndCtx(userId)
			return
		}

		bot.SendHTML(userId, utils.UserMemText(userWithMem))
	case 3:
		data, ok := state.Data.(signStudentsData)
		userIds := data.Data.Registered
		currIndex := data.Index

		if u.Message == nil || !ok || currIndex >= len(userIds) {
			bot.SendText(userId, utils.WrongMsg)
			bot.EndCtx(userId)
			return
		}

		if u.Message.Text == "Y" {
			db.Query.AddAttendance(bot.Ctx, db.AddAttendanceParams{
				UserID:   userIds[currIndex],
				LessonID: data.Data.LessonID,
				Date:     data.Data.Date,
			})
			db.Query.DecLessonsAvaliable(bot.Ctx, userIds[currIndex])
		}

		currIndex++
		if currIndex >= len(userIds) {
			bot.SendHTML(userId, fmt.Sprintf("Good job!\nThe lesson ID: <b>%d</b>", data.Data.LessonID))
			bot.EndCtx(userId)
			return
		}

		userWithMem, err := db.Query.GetUserWithMembership(bot.Ctx, userIds[currIndex])

		if err != nil {
			bot.SendText(userId, "Internal error. Text anything to continue")
			bot.Error("Sign students get user with membership err: " + err.Error())
		} else {
			bot.SendHTML(userId, utils.UserMemText(userWithMem))
		}

		data.Index = currIndex
		state.Data = data

		bot.SetCtxValue(userId, state)

		return
	}

	bot.NextCtx(userId)
}

func ChangeEmoji(bot *bot.Bot, u t.Update) {
	userId, _ := utils.UserIdFromUpdate(u)
	state, ok := bot.GetCtxValue(userId)

	if !ok {
		bot.Error(fmt.Sprintf("No scene for the user: %d", userId))
		bot.EndCtx(userId)
	}

	switch state.Stage {
	case 1:
		bot.SendHTML(userId, utils.SendEmojiMsg)
	case 2:
		if u.Message != nil {
			emoji := u.Message.Text
			isEmoji := utils.IsEmoji(emoji)

			if isEmoji {
				db.Query.UpdateEmoji(bot.Ctx, db.UpdateEmojiParams{
					ID:    userId,
					Emoji: emoji,
				})
				bot.SendText(userId, "Your new emoji: "+emoji)
			} else {
				bot.SendText(userId, "Don't make me angry. It's not an emoji 😡")
			}
		}

		bot.EndCtx(userId)
		return
	}

	bot.NextCtx(userId)
}

func AddLessons(bot *bot.Bot, u t.Update) {
	userId, _ := utils.UserIdFromUpdate(u)
	state, ok := bot.GetCtxValue(userId)

	if !ok {
		bot.Error(fmt.Sprintf("No scene for the user: %d", userId))
		bot.EndCtx(userId)
	}

	switch state.Stage {
	case 1:
		bot.SendText(userId, utils.AddLessonMsg)
		bot.NextCtx(userId)
	case 2:
		const FINISH = "Finish"

		if u.Message == nil {
			if u.CallbackQuery != nil && u.CallbackQuery.Data == FINISH {
				bot.SendText(userId, utils.GoodJob)
			} else {
				bot.SendText(userId, utils.WrongMsg)
			}

			bot.EndCtx(userId)
			return
		}

		lessonParams, err := utils.ValidateLessonInput(u.Message.Text)

		if err == nil {
			err = controller.AddLesson(bot.Ctx, lessonParams)

			if err != nil {
				bot.Error("Add lesson err: " + err.Error())
			}

			bot.SendMessage(t.Message{
				ChatId: userId,
				Text:   "New lesson was added\n\nYou can add more or or finish🧞‍♂️",
				ReplyMarkup: &t.InlineKeyboardMarkup{
					InlineKeyboard: [][]t.InlineKeyboardButton{
						{
							{
								Text:         FINISH,
								CallbackData: FINISH,
							},
						},
					},
				},
			})
		} else {
			bot.SendText(userId, "The lesson format is incorrect🔫\n"+err.Error())
			bot.EndCtx(userId)
		}
	}
}

func AssignMembership(bot *bot.Bot, u t.Update) {
	userId, _ := utils.UserIdFromUpdate(u)
	state, ok := bot.GetCtxValue(userId)

	if !ok {
		bot.Error(fmt.Sprintf("No scene for the user: %d", userId))
		bot.EndCtx(userId)
	}

	switch state.Stage {
	case 1:
		buttons := [][]t.InlineKeyboardButton{
			{
				{
					Text:         "Once a week",
					CallbackData: "1",
				},
			},
			{
				{
					Text:         "Twice a week",
					CallbackData: "2",
				},
			},
			{
				{
					Text:         "Unlimited",
					CallbackData: "8",
				},
			},
		}

		bot.SendMessage(t.Message{
			Text:   "Membership for how many days in a week?",
			ChatId: userId,
			ReplyMarkup: &t.InlineKeyboardMarkup{
				InlineKeyboard: buttons,
			},
		})
	case 2:
		if u.CallbackQuery == nil {
			bot.SendText(userId, utils.WrongMsg)
			bot.EndCtx(userId)
			return
		}

		memType, err := strconv.Atoi(u.CallbackQuery.Data)

		if err == nil {
			state.Data = memType
			bot.SetCtxValue(userId, state)

			bot.SendMessage(t.Message{
				ChatId:    userId,
				Text:      fmt.Sprintf("You choose <b>%d</b> times in a week membership\nNow, write a username or full name of a student", memType),
				ParseMode: "html",
			})
		}
	case 3:
		if u.Message == nil {
			bot.SendText(userId, utils.WrongMsg)
			bot.EndCtx(userId)
			return
		}

		shouldContinue := sendUserList(bot, userId, u.Message.Text)

		if !shouldContinue {
			return
		}
	case 4:
		if u.Message == nil {
			bot.SendText(userId, utils.WrongMsg)
			bot.EndCtx(userId)
			return
		}

		studentId, err := strconv.ParseInt(u.Message.Text, 10, 64)
		data, ok := state.Data.(int)

		if err == nil && ok {
			membership := controller.UpdateMembership(bot.Ctx, studentId, data)
			action.SendProfile(bot, studentId)
			bot.SendText(studentId, "Your membership was updated 🌋🧯")
			bot.SendText(userId, fmt.Sprintf("Gotcha🦾\nLessons avaliable: %d", membership.LessonsAvailable))
		} else {
			bot.SendText(userId, "It's not an ID🔫")
		}
		bot.EndCtx(userId)
		return
	}

	bot.NextCtx(userId)
}

func NotifyAboutLessons(bot *bot.Bot, u t.Update) {
	userId, _ := utils.UserIdFromUpdate(u)
	state, ok := bot.GetCtxValue(userId)

	if !ok {
		bot.Error(fmt.Sprintf("No scene for the user: %d", userId))
		bot.EndCtx(userId)
	}

	switch state.Stage {
	case 1:
		bot.SendMessage(t.Message{
			Text:        "Notify all users about new lessons?",
			ChatId:      userId,
			ReplyMarkup: &utils.ConformationInlineKeyboard,
		})
	case 2:
		bot.EndCtx(userId)
		if u.CallbackQuery == nil {
			bot.SendText(userId, utils.WrongMsg)
			return
		}

		if u.CallbackQuery.Data == "YES" {
			var wg sync.WaitGroup
			ids, err := db.Query.GetUsersIDs(bot.Ctx)

			if err != nil {

			}

			btns := utils.BuildInlineKeyboard([]string{utils.Timetable})

			for i := range ids {
				wg.Add(1)

				go func() {
					defer wg.Done()
					bot.SendSticker(ids[i], utils.PinkSheepMeditating)
					bot.SendMessage(t.Message{
						ChatId:      ids[i],
						Text:        "My dear student, the new timetable is waiting for you.\nSee you at the lesson🙏",
						ReplyMarkup: btns,
					})
				}()
			}
			wg.Wait()
		}
		return
	}

	bot.NextCtx(userId)
}

type freezeMemData struct {
	Type   string
	UserId int64
	Days   int
}

func FreezeMembership(bot *bot.Bot, u t.Update) {
	userId, _ := utils.UserIdFromUpdate(u)
	state, ok := bot.GetCtxValue(userId)

	if !ok {
		bot.Error(fmt.Sprintf("No scene for the user: %d", userId))
		bot.EndCtx(userId)
	}

	Single := "Single"
	All := "All"
	SendNumberDaysMsg := "Please send the <b>number</b> of days you'd like to freeze student's membership by"

	switch state.Stage {
	case 1:
		btns := utils.BuildInlineKeyboard([]string{Single, All})

		bot.SendMessage(t.Message{
			ChatId:      userId,
			Text:        "What kind of membership do you want to freeze?",
			ReplyMarkup: btns,
		})
	case 2:
		if u.CallbackQuery == nil {
			bot.SendText(userId, utils.WrongMsg)
			bot.EndCtx(userId)
			return
		}

		if u.CallbackQuery.Data == Single {
			state.Data = freezeMemData{
				Type: Single,
			}
			bot.SendText(userId, "Now, write a username or full name of a student")
		} else if u.CallbackQuery.Data == All {
			state.Stage = 4
			state.Data = freezeMemData{
				Type: All,
			}
			bot.SendHTML(userId, SendNumberDaysMsg)
		} else {
			bot.SendText(userId, utils.WrongMsg)
			bot.EndCtx(userId)
			return
		}

		bot.SetCtxValue(userId, state)
	case 3:
		if u.Message == nil {
			bot.SendText(userId, utils.WrongMsg)
			bot.EndCtx(userId)
			return
		}

		shouldContinue := sendUserList(bot, userId, u.Message.Text)

		if !shouldContinue {
			return
		}
	case 4:
		if u.Message == nil {
			bot.SendText(userId, utils.WrongMsg)
			bot.EndCtx(userId)
			return
		}

		studentId, err := strconv.ParseInt(u.Message.Text, 10, 64)
		data, ok := state.Data.(freezeMemData)

		if err == nil && ok {
			data.UserId = studentId
			state.Data = data

			bot.SetCtxValue(userId, state)

			bot.SendHTML(userId, SendNumberDaysMsg)
		} else {
			bot.SendText(userId, "It's not an ID🔫")
			bot.EndCtx(userId)
			return
		}
	case 5:
		if u.Message == nil {
			bot.SendText(userId, utils.NotANumberMsg)
			bot.EndCtx(userId)
			return
		}

		number, err := strconv.Atoi(u.Message.Text)

		if err != nil {
			bot.SendText(userId, utils.NotANumberMsg)
			bot.EndCtx(userId)
			return
		}

		data := state.Data.(freezeMemData)
		data.Days = number
		state.Data = data

		bot.SetCtxValue(userId, state)

		bot.SendMessage(t.Message{
			ChatId:      userId,
			Text:        fmt.Sprintf("Are you sure you want to freeze the membership for %d days? 🚨", number),
			ReplyMarkup: &utils.ConformationInlineKeyboard,
		})
	case 6:
		if u.CallbackQuery == nil {
			bot.SendText(userId, utils.WrongMsg)
			bot.EndCtx(userId)
			return
		}

		if u.CallbackQuery.Data == "YES" {
			var ids []int64
			var err error
			var wg sync.WaitGroup
			data := state.Data.(freezeMemData)

			if data.Type == All {
				ids, err = db.Query.GetUsersIDsWithValidMem(bot.Ctx)

				if err != nil {
					bot.EndCtx(userId)
					bot.Error("Get available lesson error: " + err.Error())
					bot.SendText(userId, "Error: "+err.Error())
					return
				}
			} else {
				ids = []int64{data.UserId}
			}

			oldUsersWithMem, err := db.Query.GetUsersWithMembership(bot.Ctx, ids)
			if err != nil {
				bot.EndCtx(userId)
				bot.Error("Get users with membership error: " + err.Error())
				bot.SendText(userId, "Error: "+err.Error())
				return
			}

			err = db.Query.AddDaysToMem(bot.Ctx, db.AddDaysToMemParams{
				Column1: data.Days,
				UserIds: ids,
			})
			if err != nil {
				bot.EndCtx(userId)
				bot.Error("Add days to mem error: " + err.Error())
				bot.SendText(userId, "Error: "+err.Error())
				return
			}

			newUsersWithMem, err := db.Query.GetUsersWithMembership(bot.Ctx, ids)
			if err != nil {
				bot.EndCtx(userId)
				bot.Error("Get users with membership error: " + err.Error())
				bot.SendText(userId, "Error: "+err.Error())
				return
			}

			oldMemMap := make(map[int64]db.GetUserWithMembershipRow)
			for _, mem := range oldUsersWithMem {
				oldMemMap[mem.ID] = utils.ConvertToUserWithMembership(mem)
			}

			newMemMap := make(map[int64]db.GetUserWithMembershipRow)
			for _, mem := range newUsersWithMem {
				newMemMap[mem.ID] = utils.ConvertToUserWithMembership(mem)
			}

			for i := range ids {
				wg.Add(1)

				go func(userID int64) {
					defer wg.Done()
					bot.SendHTML(userID, utils.NoClassesMsg)
					action.SendProfileByMem(bot, userID, oldMemMap[userID])
					bot.SendText(userID, utils.UpdatedMembershipMsg)
					action.SendProfileByMem(bot, userID, newMemMap[userID])
				}(ids[i])
			}
			wg.Wait()
		}

		bot.SendText(userId, utils.GoodJob)
		bot.EndCtx(userId)
		return
	}

	bot.NextCtx(userId)
}

func ForwardAll(bot *bot.Bot, u t.Update) {
	userID, _ := utils.UserIdFromUpdate(u)
	state, ok := bot.GetCtxValue(userID)

	if !ok {
		bot.Error(fmt.Sprintf("No scene for the user: %d", userID))
		bot.EndCtx(userID)
	}

	switch state.Stage {
	case 1:
		bot.SendHTML(userID, "Send a message that will be forwarded to everyone")
	case 2:
		if u.Message == nil {
			bot.SendText(userID, utils.WrongMsg)
			return
		}

		state.Data = u.Message.MessageID
		bot.SetCtxValue(userID, state)

		bot.SendMessage(t.Message{
			ChatId:      userID,
			Text:        "Are you sure you want to send this message?",
			ReplyMarkup: &utils.ConformationInlineKeyboard,
		})
	case 3:
		if u.CallbackQuery == nil {
			bot.SendText(userID, utils.WrongMsg)
			return
		}

		state, ok := bot.GetCtxValue(userID)
		if !ok {
			bot.SendText(userID, utils.WrongMsg)
			return
		}

		messageID, ok := state.Data.(int)
		if !ok {
			bot.SendText(userID, utils.WrongMsg)
			return
		}

		bot.EndCtx(userID)

		if u.CallbackQuery.Data == "YES" {
			var idsToBlock []int64
			ids, err := db.Query.GetUsersIDs(bot.Ctx)

			if err != nil {
				bot.SendText(userID, err.Error())
				return
			}

			bot.SendText(userID, "Starting forwarding, please wait🗿")
			for i := range ids {
				_, err := bot.Forward(ids[i], userID, messageID)

				if err == t.BotIsBlockedError {
					idsToBlock = append(idsToBlock, ids[i])
				}
			}
			bot.SendText(userID, "Success🐑")
			err = db.Query.BlockUsers(bot.Ctx, idsToBlock)

			if err != nil {
				bot.SendText(userID, "BlockUsers err: "+err.Error())
				return
			}

			if len(idsToBlock) > 0 {
				bot.SendText(userID, fmt.Sprintf("Great success, %d blocked yoga bot", len(idsToBlock)))
			}
		} else {
			bot.SendText(userID, "Ok")
		}
		return
	}

	bot.NextCtx(userID)
}

type editLessonData struct {
	Type     string
	LessonId int
}

func EditLesson(bot *bot.Bot, u t.Update) {
	userId, _ := utils.UserIdFromUpdate(u)
	state, ok := bot.GetCtxValue(userId)

	if !ok {
		bot.Error(fmt.Sprintf("No scene for the user: %d", userId))
		bot.EndCtx(userId)
	}

	DATE := "date"
	TIME := "time"
	DESCRIPTION := "description"
	MAX := "max"

	switch state.Stage {
	case 1:
		lessons, err := db.Query.GetAvailableLessons(bot.Ctx)

		if err != nil {
			bot.EndCtx(userId)
			bot.Error("Get available lesson error: " + err.Error())
			bot.SendText(userId, utils.WrongMsg)

			return
		}

		msg := utils.GenerateTimetableMsg(lessons, false)
		msg.ChatId = userId
		msg.Text = "Which lesson you want to edit?"
		bot.SendMessage(msg)
	case 2:
		if u.CallbackQuery == nil {
			bot.SendText(userId, utils.WrongMsg)
			bot.EndCtx(userId)
			return
		}

		data := u.CallbackQuery.Data

		if !utils.LessonRegexp().MatchString(data) {
			bot.SendText(userId, utils.WrongMsg)
			bot.EndCtx(userId)
			return
		}

		lessonAndId := strings.Split(data, "=")
		lessonId, err := strconv.Atoi(lessonAndId[1])

		if err != nil {
			bot.SendText(userId, utils.WrongMsg)
			bot.EndCtx(userId)
			return
		}

		state.Data = editLessonData{
			LessonId: lessonId,
		}
		bot.SetCtxValue(userId, state)

		btns := utils.BuildInlineKeyboard([]string{
			DATE,
			TIME,
			MAX,
			DESCRIPTION,
		})

		bot.SendMessage(t.Message{
			ChatId:      userId,
			Text:        "What do you want to edit?",
			ReplyMarkup: btns,
		})
	case 3:
		if u.CallbackQuery == nil {
			bot.SendText(userId, utils.WrongMsg)
			bot.EndCtx(userId)
			return
		}

		stateData := state.Data.(editLessonData)
		stateData.Type = u.CallbackQuery.Data
		state.Data = stateData
		bot.SetCtxValue(userId, state)
		text := ""

		switch stateData.Type {
		case DATE:
			text = "Send a <b>date</b> in the format YYYY-MM-DD"
		case TIME:
			text = "Send a <b>time</b> in the format HH:MM"
		case DESCRIPTION:
			text = "Send a <b>description</b> text"
		case MAX:
			text = "Send a <b>max membership</b> number"
		}

		bot.SendHTML(userId, text)
	case 4:
		if u.Message == nil {
			bot.SendText(userId, utils.WrongMsg)
			return
		}

		isValidValue := false
		value := u.Message.Text
		lessonData := state.Data.(editLessonData)

		switch lessonData.Type {
		case DATE:
			date, err := time.Parse("2006-01-02", value)

			if err == nil {
				isValidValue = true
				db.Query.UpdateLessonDate(bot.Ctx, db.UpdateLessonDateParams{
					ID:   lessonData.LessonId,
					Date: date,
				})
			}
		case TIME:
			parsedTime, err := time.Parse("15:04", value)

			if err == nil {
				isValidValue = true
				db.Query.UpdateLessonTime(bot.Ctx, db.UpdateLessonTimeParams{
					ID:   lessonData.LessonId,
					Time: parsedTime,
				})
			}
		case DESCRIPTION:
			if len(value) > 0 {
				isValidValue = true
				db.Query.UpdateLessonDesc(bot.Ctx, db.UpdateLessonDescParams{
					ID:          lessonData.LessonId,
					Description: value,
				})
			}
		case MAX:
			maxInt, err := strconv.Atoi(value)

			if err == nil {
				isValidValue = true
				db.Query.UpdateLessonMax(bot.Ctx, db.UpdateLessonMaxParams{
					ID:  lessonData.LessonId,
					Max: maxInt,
				})
			}
		}

		if isValidValue {
			bot.SendText(userId, utils.GoodJob)
		} else {
			bot.SendText(userId, utils.WrongMsg)
		}

		bot.EndCtx(userId)
		return
	}

	bot.NextCtx(userId)
}

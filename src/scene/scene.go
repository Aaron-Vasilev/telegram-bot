package scene

import (
	"bot/src/action"
	"bot/src/bot"
	"bot/src/controller"
	"bot/src/utils"
	t "bot/src/utils/types"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

type sceneCallback = func(ctx *Ctx, bot *bot.Bot, db *sql.DB, u t.Update)

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

func Start(ctx *Ctx, bot *bot.Bot, db *sql.DB, u t.Update, scene string) {
	ctx.SetValue(u.FromChat().ID, SceneState{
		Scene: scene,
		Stage: 1,
	})

	Map[scene](ctx, bot, db, u)
}

type signStudentsData struct {
	Data  t.RegisterdOnLesson
	Index int
}

func SignStudents(ctx *Ctx, bot *bot.Bot, db *sql.DB, u t.Update) {

	userId, _ := utils.UserIdFromUpdate(u)
	state, ok := ctx.GetValue(userId)

	if !ok {
		bot.Error(fmt.Sprintf("No scene for the user: %d", userId))
		ctx.End(userId)
	}

	switch state.Stage {
	case 1:
		lessons := controller.GetAvaliableLessons(db)

		msg := utils.GenerateTimetableMsg(lessons, true)
		msg.ChatId = userId
		msg.Text = "Send me back an <b>ID</b> of a lesson"
		msg.ParseMode = "html"
		bot.SendMessage(msg)
	case 2:
		if u.Message == nil {
			bot.SendText(userId, "It's not a ID")
			ctx.End(userId)
			return
		}

		//TODO receive a lessonId also from CallbackQuery
		lessonId, err := strconv.Atoi(u.Message.Text)

		if err != nil {
			bot.SendText(userId, "The ID is not correct")
			ctx.End(userId)
			return
		}

		wasLessonSigned := controller.IsLessonSigned(db, lessonId)

		if wasLessonSigned {
			bot.SendText(userId, "This lesson was signed⚠️")
		}

		registered := controller.GetRegisteredOnLesson(db, lessonId)

		if len(registered.IDs) == 0 {
			bot.SendText(userId, "The are no users on this lesson")
			ctx.End(userId)
			return
		}
		state.Data = signStudentsData{
			Data:  registered,
			Index: 0,
		}

		ctx.SetValue(userId, state)

		userWithMem := controller.GetUserWithMembership(db, registered.IDs[0])

		bot.SendHTML(userId, utils.UserMemText(userWithMem))
	case 3:
		data, ok := state.Data.(signStudentsData)
		ids := data.Data.IDs
		currIndex := data.Index

		if u.Message == nil || !ok || currIndex >= len(ids) {
			bot.SendText(userId, utils.WrongMsg)
			ctx.End(userId)
			return
		}

		if u.Message.Text == "Y" {
			db.Exec(`INSERT INTO yoga.attendance 
                    (user_id, lesson_id, date) VALUES ($1, $2, $3);`,
				ids[currIndex], data.Data.LessonId, data.Data.Date)
			db.Exec(`UPDATE yoga.membership 
                    SET lessons_avaliable = lessons_avaliable - 1
                    WHERE user_id=$1;`, ids[currIndex])
		}

		currIndex++
		if currIndex >= len(ids) {
			bot.SendText(userId, "Good job!")
			ctx.End(userId)
			return
		}

		userWithMem := controller.GetUserWithMembership(db, ids[currIndex])
		bot.SendHTML(userId, utils.UserMemText(userWithMem))

		data.Index = currIndex
		state.Data = data

		ctx.SetValue(userId, state)

		return
	}

	ctx.Next(userId)
}

func ChangeEmoji(ctx *Ctx, bot *bot.Bot, db *sql.DB, u t.Update) {
	userId, _ := utils.UserIdFromUpdate(u)
	state, ok := ctx.GetValue(userId)

	if !ok {
		bot.Error(fmt.Sprintf("No scene for the user: %d", userId))
		ctx.End(userId)
	}

	switch state.Stage {
	case 1:
		bot.SendHTML(userId, utils.SendEmojiMsg)
	case 2:
		if u.Message != nil {
			emoji := u.Message.Text
			isEmoji := utils.IsEmoji(emoji)

			if isEmoji {
				controller.UpdateEmoji(db, userId, emoji)
				bot.SendText(userId, "Your new emoji: "+emoji)
			} else {
				bot.SendText(userId, "Don't make me angry. It's not an emoji 😡")
			}
		}

		ctx.End(userId)
		return
	}

	ctx.Next(userId)
}

func AddLessons(ctx *Ctx, bot *bot.Bot, db *sql.DB, u t.Update) {
	userId, _ := utils.UserIdFromUpdate(u)
	state, ok := ctx.GetValue(userId)

	if !ok {
		bot.Error(fmt.Sprintf("No scene for the user: %d", userId))
		ctx.End(userId)
	}

	switch state.Stage {
	case 1:
		bot.SendText(userId, utils.AddLessonMsg)
		ctx.Next(userId)
	case 2:
		const FINISH = "Finish"

		if u.Message == nil {
			if u.CallbackQuery != nil && u.CallbackQuery.Data == FINISH {
				bot.SendText(userId, utils.GoodJob)
			} else {
				bot.SendText(userId, utils.WrongMsg)
			}

			ctx.End(userId)
			return
		}

		data := utils.ValidateLessonStr(u.Message.Text)

		if data.IsValid {
			controller.AddLesson(db, data)
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
			bot.SendText(userId, "The lesson format is incorrect🔫")
			ctx.End(userId)
		}
	}
}

func AssignMembership(ctx *Ctx, bot *bot.Bot, db *sql.DB, u t.Update) {
	userId, _ := utils.UserIdFromUpdate(u)
	state, ok := ctx.GetValue(userId)

	if !ok {
		bot.Error(fmt.Sprintf("No scene for the user: %d", userId))
		ctx.End(userId)
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
			ctx.End(userId)
			return
		}

		memType, err := strconv.Atoi(u.CallbackQuery.Data)

		if err == nil {
			state.Data = memType
			ctx.SetValue(userId, state)

			bot.SendMessage(t.Message{
				ChatId:    userId,
				Text:      fmt.Sprintf("You choose <b>%d</b> times in a week membership\nNow, write a username or full name of a student", memType),
				ParseMode: "html",
			})
		}
	case 3:
		if u.Message == nil {
			bot.SendText(userId, utils.WrongMsg)
			ctx.End(userId)
			return
		}

		sendUserList(ctx, db, bot, userId, u.Message.Text)
	case 4:
		if u.Message == nil {
			bot.SendText(userId, utils.WrongMsg)
			ctx.End(userId)
			return
		}

		studentId, err := strconv.ParseInt(u.Message.Text, 10, 64)
		data, ok := state.Data.(int)

		if err == nil && ok {
			membership := controller.UpdateMembership(db, studentId, data)

			action.SendProfile(bot, db, studentId)
			bot.SendText(studentId, "Your membership was updated 🌋🧯")
			bot.SendText(userId, fmt.Sprintf("Gotcha🦾\nLessons avaliable: %d", membership.LessonsAvailable))
		} else {
			bot.SendText(userId, "It's not an ID🔫")
		}
		ctx.End(userId)
		return
	}

	ctx.Next(userId)
}

func NotifyAboutLessons(ctx *Ctx, bot *bot.Bot, db *sql.DB, u t.Update) {
	userId, _ := utils.UserIdFromUpdate(u)
	state, ok := ctx.GetValue(userId)

	if !ok {
		bot.Error(fmt.Sprintf("No scene for the user: %d", userId))
		ctx.End(userId)
	}

	switch state.Stage {
	case 1:
		bot.SendMessage(t.Message{
			Text:        "Notify all users about new lessons?",
			ChatId:      userId,
			ReplyMarkup: &utils.ConformationInlineKeyboard,
		})
	case 2:
		ctx.End(userId)
		if u.CallbackQuery == nil {
			bot.SendText(userId, utils.WrongMsg)
			return
		}

		if u.CallbackQuery.Data == "YES" {
			var wg sync.WaitGroup
			ids := controller.GetUsersIDs(db)
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

	ctx.Next(userId)
}

type freezeMemData struct {
	Type   string
	UserId int64
	Days   int
}

func FreezeMembership(ctx *Ctx, bot *bot.Bot, db *sql.DB, u t.Update) {
	userId, _ := utils.UserIdFromUpdate(u)
	state, ok := ctx.GetValue(userId)

	if !ok {
		bot.Error(fmt.Sprintf("No scene for the user: %d", userId))
		ctx.End(userId)
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
			ctx.End(userId)
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
			ctx.End(userId)
			return
		}

		ctx.SetValue(userId, state)
	case 3:
		if u.Message == nil {
			bot.SendText(userId, utils.WrongMsg)
			ctx.End(userId)
			return
		}

		sendUserList(ctx, db, bot, userId, u.Message.Text)
	case 4:
		if u.Message == nil {
			bot.SendText(userId, utils.WrongMsg)
			ctx.End(userId)
			return
		}

		studentId, err := strconv.ParseInt(u.Message.Text, 10, 64)
		data, ok := state.Data.(freezeMemData)

		if err == nil && ok {
			data.UserId = studentId
			state.Data = data

			ctx.SetValue(userId, state)

			bot.SendHTML(userId, SendNumberDaysMsg)
		} else {
			bot.SendText(userId, "It's not an ID🔫")
			ctx.End(userId)
			return
		}
	case 5:
		if u.Message == nil {
			bot.SendText(userId, utils.NotANumberMsg)
			ctx.End(userId)
			return
		}

		number, err := strconv.Atoi(u.Message.Text)

		if err != nil {
			bot.SendText(userId, utils.NotANumberMsg)
			ctx.End(userId)
			return
		}

		data := state.Data.(freezeMemData)
		data.Days = number
		state.Data = data

		ctx.SetValue(userId, state)

		bot.SendMessage(t.Message{
			ChatId:      userId,
			Text:        fmt.Sprintf("Are you sure you want to freeze the membership for %d days? 🚨", number),
			ReplyMarkup: &utils.ConformationInlineKeyboard,
		})
	case 6:
		if u.CallbackQuery == nil {
			bot.SendText(userId, utils.WrongMsg)
			ctx.End(userId)
			return
		}

		if u.CallbackQuery.Data == "YES" {
			var ids []int64
			var wg sync.WaitGroup
			data := state.Data.(freezeMemData)

			if data.Type == All {
				ids = controller.GetUsersIDsWithValidMem(db)
			} else {
				ids = []int64{data.UserId}
			}

			for i := range ids {
				wg.Add(1)

				go func() {
					defer wg.Done()
					bot.SendHTML(ids[i], utils.NoClassesMsg)
					action.SendProfile(bot, db, ids[i])
					controller.AddDaysToMem(db, ids[i], data.Days)
					bot.SendText(ids[i], utils.UpdatedMembershipMsg)
					action.SendProfile(bot, db, ids[i])
				}()
			}
			wg.Wait()
		}

		bot.SendText(userId, utils.GoodJob)
		ctx.End(userId)
		return
	}

	ctx.Next(userId)
}

func ForwardAll(ctx *Ctx, bot *bot.Bot, db *sql.DB, u t.Update) {
	userID, _ := utils.UserIdFromUpdate(u)
	state, ok := ctx.GetValue(userID)

	if !ok {
		bot.Error(fmt.Sprintf("No scene for the user: %d", userID))
		ctx.End(userID)
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
		ctx.SetValue(userID, state)

		bot.SendMessage(t.Message{
			ChatId:      userID,
			Text:        "Are you sure you want to send this message?",
			ReplyMarkup: &utils.ConformationInlineKeyboard,
		})
	case 3:
		if u.CallbackQuery == nil {
			bot.SendText(userID, utils.WrongMsg)
			ctx.End(userID)
			return
		}

		state, ok := ctx.GetValue(userID)
		if !ok {
			bot.SendText(userID, utils.WrongMsg)
			ctx.End(userID)
			return
		}

		messageID, ok := state.Data.(int)
		if !ok {
			bot.SendText(userID, utils.WrongMsg)
			ctx.End(userID)
			return
		}

		if u.CallbackQuery.Data == "YES" {
			var idsToBlock []int64
			ids := controller.GetUsersIDs(db)

			for i := range ids {
				_, err := bot.Forward(ids[i], userID, messageID)

				if err == t.BotIsBlockedError {
					idsToBlock = append(idsToBlock, ids[i])
				}
			}
			err := controller.BlockUsers(db, idsToBlock)

			if err != nil {
				bot.Error("Forward all err: " + err.Error())
			}
			bot.SendText(userID, fmt.Sprintf("Great success, %d blocked yoga bot", len(idsToBlock)))
		} else {
			bot.SendText(userID, "Ok")
		}
		ctx.End(userID)
		return
	}

	ctx.Next(userID)
}

type editLessonData struct {
	Type     string
	LessonId int
}

func EditLesson(ctx *Ctx, bot *bot.Bot, db *sql.DB, u t.Update) {
	userId, _ := utils.UserIdFromUpdate(u)
	state, ok := ctx.GetValue(userId)

	if !ok {
		bot.Error(fmt.Sprintf("No scene for the user: %d", userId))
		ctx.End(userId)
	}

	date := "date"
	time := "time"
	description := "description"
	max := "max"

	switch state.Stage {
	case 1:
		lessons := controller.GetAvaliableLessons(db)

		msg := utils.GenerateTimetableMsg(lessons, false)
		msg.ChatId = userId
		msg.Text = "Which lesson you want to edit?"
		bot.SendMessage(msg)
	case 2:
		if u.CallbackQuery == nil {
			bot.SendText(userId, utils.WrongMsg)
			ctx.End(userId)
			return
		}

		data := u.CallbackQuery.Data

		if !utils.LessonRegexp().MatchString(data) {
			bot.SendText(userId, utils.WrongMsg)
			ctx.End(userId)
			return
		}

		lessonAndId := strings.Split(data, "=")
		lessonId, err := strconv.Atoi(lessonAndId[1])

		if err != nil {
			bot.SendText(userId, utils.WrongMsg)
			ctx.End(userId)
			return
		}

		state.Data = editLessonData{
			LessonId: lessonId,
		}
		ctx.SetValue(userId, state)

		btns := utils.BuildInlineKeyboard([]string{
			date,
			time,
			max,
			description,
		})

		bot.SendMessage(t.Message{
			ChatId:      userId,
			Text:        "What do you want to edit?",
			ReplyMarkup: btns,
		})
	case 3:
		if u.CallbackQuery == nil {
			bot.SendText(userId, utils.WrongMsg)
			ctx.End(userId)
			return
		}

		stateData := state.Data.(editLessonData)
		stateData.Type = u.CallbackQuery.Data
		state.Data = stateData
		ctx.SetValue(userId, state)
		text := ""

		switch stateData.Type {
		case date:
			text = "Send a <b>date</b> in the format YYYY-MM-DD"
		case time:
			text = "Send a <b>time</b> in the format HH:MM"
		case description:
			text = "Send a <b>description</b> text"
		case max:
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
		case date:
			if utils.DateRegexp().Match([]byte(value)) {
				isValidValue = true
				controller.UpdateLessonDate(db, lessonData.LessonId, value)
			}
		case time:
			if utils.TimeRegexp().Match([]byte(value)) {
				isValidValue = true
				controller.UpdateLessonTime(db, lessonData.LessonId, value)
			}
		case description:
			if len(value) > 0 {
				isValidValue = true
				controller.UpdateLessonDesc(db, lessonData.LessonId, value)
			}
		case max:
			maxInt, err := strconv.Atoi(value)

			if err == nil {
				isValidValue = true
				controller.UpdateLessonMax(db, lessonData.LessonId, maxInt)
			}
		}

		if isValidValue {
			bot.SendText(userId, utils.GoodJob)
		} else {
			bot.SendText(userId, utils.WrongMsg)
		}

		ctx.End(userId)
		return
	}

	ctx.Next(userId)
}

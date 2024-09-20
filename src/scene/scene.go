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
	utils.ExtendMemDate:      ExtendMemEndDate,
}

func Start(ctx *Ctx, bot *bot.Bot, db *sql.DB, u t.Update, scene string) {
	ctx.SetValue(u.FromChat().ID, SceneState{
		Scene: scene,
		Stage: 1,
	})

	Map[scene](ctx, bot, db, u)
}

func SignStudents(ctx *Ctx, bot *bot.Bot, db *sql.DB, u t.Update) {
	type signStudentsData struct {
		Data  t.RegisterdOnLesson
		Index int
	}

	userId, _ := utils.UserIdFromUpdate(u)
	state, ok := ctx.GetValue(userId)

	if !ok {
		bot.Error(fmt.Sprintf("No scene for the user: %d", userId))
		ctx.End(userId)
	}

	switch state.Stage {
	case 1:
		lessons := controller.GetAvaliableLessons(db)

		msg := utils.GenerateTimetable(lessons, true)
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

		//TODO receive a and lessonId from CallbackQuery
		lessonId, err := strconv.Atoi(u.Message.Text)

		if err != nil {
			bot.SendText(userId, "The ID is not correct")
			ctx.End(userId)
			return
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

		data := utils.ValidateLessonMsg(u.Message.Text)

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

		users := controller.FindUsersByName(db, u.Message.Text)

		if len(users) == 0 {
			bot.SendText(userId, "There are no users like: "+u.Message.Text)
			ctx.End(userId)
			return
		}

		for i := range users {
			userName := ""

			if users[i].Username.Valid {
				userName = "@" + users[i].Username.String
			}
			bot.SendText(userId, fmt.Sprintf("%s %s ID = %d", users[i].Name, userName, users[i].ID))
		}
		bot.SendText(userId, "Send back the ID of the user you want to assign a membership")
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

			for i := range ids {
				wg.Add(1)

				go func() {
					defer wg.Done()
					bot.SendSticker(ids[i], utils.PinkSheepMeditating)
					bot.SendMessage(t.Message{
						ChatId: ids[i],
						Text:   "My dear student, the new timetable is waiting for you.\nSee you at the lesson🙏",
						ReplyMarkup: &t.InlineKeyboardMarkup{
							InlineKeyboard: [][]t.InlineKeyboardButton{
								{
									{
										Text:         utils.Timetable,
										CallbackData: utils.Timetable,
									},
								},
							},
						},
					})
				}()
			}
			wg.Wait()
		}
		return
	}

	ctx.Next(userId)
}

func ExtendMemEndDate(ctx *Ctx, bot *bot.Bot, db *sql.DB, u t.Update) {
	userId, _ := utils.UserIdFromUpdate(u)
	state, ok := ctx.GetValue(userId)

	if !ok {
		bot.Error(fmt.Sprintf("No scene for the user: %d", userId))
		ctx.End(userId)
	}

	switch state.Stage {
	case 1:
		bot.SendHTML(userId, "Please send the <b>number</b> of days you'd like to extend student's membership by")
	case 2:
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
		state.Data = number
		ctx.SetValue(userId, state)

		bot.SendMessage(t.Message{
			ChatId:      userId,
			Text:        fmt.Sprintf("Are you sure you want to extend each membership for %d days?", number),
			ReplyMarkup: &utils.ConformationInlineKeyboard,
		})
	case 3:
		if u.CallbackQuery == nil {
			bot.SendText(userId, utils.WrongMsg)
			ctx.End(userId)
			return
		}

		state, ok := ctx.GetValue(userId)
		if !ok {
			bot.SendText(userId, utils.WrongMsg)
			ctx.End(userId)
			return
		}

		daysNum, ok := state.Data.(int)
		if !ok {
			bot.SendText(userId, utils.WrongMsg)
			ctx.End(userId)
			return
		}

		if u.CallbackQuery.Data == "YES" {
			var wg sync.WaitGroup
			ids := controller.GetUsersIDsWithValidMem(db)

			for i := range ids {
				wg.Add(1)

				go func() {
					defer wg.Done()
					bot.SendHTML(ids[i], utils.NoClassesMsg)
					action.SendProfile(bot, db, ids[i])
					controller.AddDaysToMem(db, ids[i], daysNum)
					bot.SendText(ids[i], utils.UpdatedMembershipMsg)
					action.SendProfile(bot, db, ids[i])
				}()
			}
			wg.Wait()
		}
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
			ids := controller.GetUsersIDs(db)

			for i := range ids {
				go bot.Forward(ids[i], userID, messageID)
			}
		}
		ctx.End(userID)
		return
	}

	ctx.Next(userID)
}

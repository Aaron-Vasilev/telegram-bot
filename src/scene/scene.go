package scene

import (
	"bot/src/bot"
	"bot/src/controller"
	"bot/src/utils"
	t "bot/src/utils/types"
	"database/sql"
	"fmt"
	"strconv"
	"sync"
)

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
		bot.SendMessage(msg)
	case 2:
		if u.Message == nil {
			bot.SendText(userId, utils.WrongMsg)
			ctx.End(userId)
			return
		}

		//TODO receive a and lessonId from CallbackQuery
		lessonId, err := strconv.Atoi(u.Message.Text)

		if err != nil {
			bot.SendText(userId, utils.WrongMsg)
			ctx.End(userId)
			return
		}

		registered := controller.GetRegisteredOnLesson(db, lessonId)
		state.Data = signStudentsData{
			Data:  registered,
			Index: 0,
		}

		ctx.SetValue(userId, state)

		userWithMem := controller.GetUserWithMembership(db, userId)

		bot.SendHTML(userId, utils.UserMemText(userWithMem))
	case 3:
		data, ok := state.Data.(signStudentsData)
		registered := data.Data.IDs
		currIndex := data.Index

		if u.Message == nil || !ok || currIndex >= len(registered) {
			bot.SendText(userId, utils.WrongMsg)
			ctx.End(userId)
			return
		}

		registeredID := registered[currIndex]

		if u.Message.Text == "Y" {
			db.Exec(`INSERT INTO yoga.attendance 
                    (user_id, lesson_id, date) VALUES ($1, $2, $3);`,
				registeredID, data.Data.LessonId, data.Data.Date)
			db.Exec(`UPDATE yoga.membership 
                    SET lessons_avaliable = lessons_avaliable - 1
                    WHERE user_id=$1;`, registeredID)
		}

		currIndex++
		if currIndex >= len(registered) {
			bot.SendText(userId, "Good job!")
			ctx.End(userId)
			return
		}

		userWithMem := controller.GetUserWithMembership(db, userId)
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
		bot.SendText(userId, utils.SendEmojiMsg)
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
		if u.Message == nil {
			ctx.End(userId)
			return
		}

		data := utils.ValidateLessonMsg(u.Message.Text)

		if data.IsValid {
			controller.AddLesson(db, data)
			bot.SendText(userId, "New lesson is added\n\nYou can add more or leave it as it is🧞‍♂️")
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

		for i := range users {
			bot.SendText(userId, fmt.Sprintf("%s @%s ID = %d", users[i].Name, users[i].Username, users[i].ID))
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

			bot.SendText(userId, fmt.Sprintf("Gotcha🦾\nLessons avaliable: %d", membership.LessonsAvailable))
		} else {
			bot.SendText(userId, "It's not an ID🔫")
		}
		ctx.End(userId)
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
			Text:   "Notify all users about new lessons?",
			ChatId: userId,
			ReplyMarkup: &t.InlineKeyboardMarkup{
				InlineKeyboard: [][]t.InlineKeyboardButton{
					{
						{
							Text:         "Yes",
							CallbackData: "YES",
						},
						{
							Text:         "No",
							CallbackData: "NO",
						},
					},
				},
			},
		})
	case 2:
		if u.CallbackQuery == nil {
			bot.SendText(userId, utils.WrongMsg)
			ctx.End(userId)
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
		ctx.End(userId)
		return
	}

	ctx.Next(userId)
}

func notifyUsers(bot *bot.Bot, ids []int64, msgs ...t.Message) {
}

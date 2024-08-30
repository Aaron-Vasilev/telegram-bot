package scene

import (
	"bot/src/bot"
	"bot/src/controller"
	"bot/src/utils"
	t "bot/src/utils/types"
	"database/sql"
	"fmt"
	"strconv"
)

func GenTokenScene(ctx *Ctx, bot *bot.Bot, db *sql.DB, u t.Update) {
	userId, _ := utils.UserIdFromUpdate(u) 
	state, ok := ctx.GetValue(userId)

	if !ok {
		bot.Error(fmt.Sprintf("No scene for the user: %d",  userId))
		ctx.End(userId)
	}

	switch state.Stage {
	case 1:
		buttons := [][]t.InlineKeyboardButton{
			{
				{
					Text: "Once a week",
					CallbackData: "1",
				},
			},
		}

		bot.SendMessage(t.Message {
			Text: "Membership for how many days in a week?",
			ChatId: userId,
			ReplyMarkup: &t.InlineKeyboardMarkup{
				InlineKeyboard: buttons,
			},

		})
	case 2:
		if u.CallbackQuery != nil {
			uuidStr := controller.CreateToken(db, u.CallbackQuery.Data)

			bot.SendText(userId, uuidStr)
		}

		ctx.End(userId)
		return 
	}

	ctx.Next(userId)
}


func ChangeEmoji(ctx *Ctx, bot *bot.Bot, db *sql.DB, u t.Update) {
	userId, _ := utils.UserIdFromUpdate(u) 
	state, ok := ctx.GetValue(userId)

	if !ok {
		bot.Error(fmt.Sprintf("No scene for the user: %d",  userId))
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
				bot.SendText(userId, "Your new emoji: " + emoji)
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
		bot.Error(fmt.Sprintf("No scene for the user: %d",  userId))
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
		bot.Error(fmt.Sprintf("No scene for the user: %d",  userId))
		ctx.End(userId)
	}

	switch state.Stage {
	case 1:
		buttons := [][]t.InlineKeyboardButton{
			{
				{
					Text: "Once a week",
					CallbackData: "1",
				},
			},
		}

		bot.SendMessage(t.Message {
			Text: "Membership for how many days in a week?",
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
			ChatId: userId,
			Text: fmt.Sprintf("You choose <b>%d</b> times in a week membership\nNow, write a username or full name of a student", memType),
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

		for i := range(users) {
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

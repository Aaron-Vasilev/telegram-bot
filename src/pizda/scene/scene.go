package scene

import (
	"bot/src/bot"
	"bot/src/common"
	"bot/src/pizda/db"
	cnst "bot/src/pizda/utils/const"
	"bot/src/utils"
	yogaUtils "bot/src/utils"
	t "bot/src/utils/types"
	"fmt"
	"strconv"
)

func RegisterScenes(bot *bot.Bot) {
	bot.RegisterScene(cnst.ForwardAll, ForwardAll)
	bot.RegisterScene(cnst.AssignSubscription, assignSubscription)
	bot.RegisterScene(cnst.ExtendPayment, extendPayment)
}

func ForwardAll(bot *bot.Bot, u t.Update) {
	userID, _ := yogaUtils.UserIdFromUpdate(u)
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
			bot.SendText(userID, yogaUtils.WrongMsg)
			return
		}

		state.Data = u.Message.MessageID
		bot.SetCtxValue(userID, state)

		bot.SendMessage(t.Message{
			ChatId:      userID,
			Text:        "Are you sure you want to send this message?",
			ReplyMarkup: &yogaUtils.ConformationInlineKeyboard,
		})
	case 3:
		if u.CallbackQuery == nil {
			bot.SendText(userID, yogaUtils.WrongMsg)
			bot.EndCtx(userID)
			return
		}

		state, ok := bot.GetCtxValue(userID)
		if !ok {
			bot.SendText(userID, yogaUtils.WrongMsg)
			bot.EndCtx(userID)
			return
		}

		messageID, ok := state.Data.(int)
		if !ok {
			bot.SendText(userID, yogaUtils.WrongMsg)
			bot.EndCtx(userID)
			return
		}

		if u.CallbackQuery.Data == "YES" {
			var idsToBlock []int64
			ids, err := db.Query.GetUsersIDs(bot.Ctx)

			if err != nil {
				bot.SendText(userID, err.Error())
				bot.EndCtx(userID)
				return
			}

			for i := range ids {
				_, err := bot.Forward(ids[i], userID, messageID)

				if err == t.BotIsBlockedError {
					idsToBlock = append(idsToBlock, ids[i])
				}
			}
			err = db.Query.BlockUsers(bot.Ctx, idsToBlock)

			if err != nil {
				bot.SendText(userID, "BlockUsers err: "+err.Error())
				bot.EndCtx(userID)
				return
			}

			bot.SendText(userID, fmt.Sprintf("Great success, %d blocked yoga bot", len(idsToBlock)))
		} else {
			bot.SendText(userID, "Ok")
		}
		bot.EndCtx(userID)
		return
	}

	bot.NextCtx(userID)
}

func assignSubscription(bot *bot.Bot, u t.Update) {
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
					Text:         "🇮🇱  Bit, Hapoalim",
					CallbackData: string(db.PizdaPaymentMethodBIT),
				},
			},
			{
				{
					Text:         "🇷🇺 Tinkoff",
					CallbackData: string(db.PizdaPaymentMethodMIR),
				},
			},
		}

		bot.SendMessage(t.Message{
			Text:   "Оплата была произведена по",
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

		state.Data = u.CallbackQuery.Data
		bot.SetCtxValue(userId, state)

		bot.SendMessage(t.Message{
			ChatId:    userId,
			Text:      "Пришил мне ник, фамилию или имя пользователя",
			ParseMode: "html",
		})
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

		payerId, err := strconv.ParseInt(u.Message.Text, 10, 64)
		var method db.PizdaPaymentMethod
		ok := false
		if str, isStr := state.Data.(string); isStr {
			method = db.PizdaPaymentMethod(str)
			ok = true
		}

		if err == nil && ok {
			err := db.Query.AddPayment(bot.Ctx, db.AddPaymentParams{
				UserID: payerId,
				Method: method,
			})

			if err == nil {
				bot.SendText(userId, "Success 🌋🧯")
				bot.SendMessage(
					common.GenerateKeyboardMsg(
						payerId,
						cnst.PayKeyboard,
						"Доспут к курcу получен, можешь приступать к практикам🙏\nTы ещё скажешь себе за это спасибо, а пока, спасибо тебе🥰",
					),
				)
			} else {
				bot.SendText(userId, utils.WrongMsg)
			}
		} else {
			bot.SendText(userId, "It's not an ID🔫")
		}

		bot.EndCtx(userId)
		return
	}

	bot.NextCtx(userId)
}

func extendPayment(bot *bot.Bot, u t.Update) {
	userId, _ := utils.UserIdFromUpdate(u)
	state, ok := bot.GetCtxValue(userId)

	if !ok {
		bot.Error(fmt.Sprintf("No scene for the user: %d", userId))
		bot.EndCtx(userId)
		return
	}

	switch state.Stage {
	case 1:
		bot.SendMessage(t.Message{
			ChatId:    userId,
			Text:      "Пришли мне ник, фамилию или имя пользователя",
			ParseMode: "html",
		})
	case 2:
		if u.Message == nil {
			bot.SendText(userId, utils.WrongMsg)
			bot.EndCtx(userId)
			return
		}

		shouldContinue := sendUserList(bot, userId, u.Message.Text)

		if !shouldContinue {
			return
		}
	case 3:
		if u.Message == nil {
			bot.SendText(userId, utils.WrongMsg)
			bot.EndCtx(userId)
			return
		}

		userId64, err := strconv.ParseInt(u.Message.Text, 10, 64)

		if err != nil {
			bot.SendText(userId, "It's not an ID🔫")
			bot.EndCtx(userId)
			return
		}

		payment, err := db.Query.GetValidPayment(bot.Ctx, userId64)

		if err != nil {
			bot.SendText(userId, "User doesn't have an active payment")
			bot.EndCtx(userId)
			return
		}

		err = db.Query.ExtendPaymentByMonth(bot.Ctx, payment.ID)

		if err != nil {
			bot.SendText(userId, "Error extending payment: "+err.Error())
			bot.EndCtx(userId)
			return
		}

		bot.SendText(userId, "Payment extended by 1 month 🎉")
		bot.SendText(userId64, "Твоя подписка была продлена на месяц! 🎊")
		bot.EndCtx(userId)
		return
	}

	bot.NextCtx(userId)
}

func sendUserList(bot *bot.Bot, userID int64, search string) bool {
	shouldContinue := true
	textWithUsers := ""
	users, err := db.Query.FindUsersByName(bot.Ctx, search)

	if err != nil {
		bot.Error("find users by name error:" + err.Error())
	}

	if len(users) == 0 {
		bot.SendText(userID, "There are no users like: "+search)
		bot.EndCtx(userID)

		shouldContinue = false
	} else {
		for i := range users {
			userName := ""

			if users[i].Username != "" {
				userName = "@" + users[i].Username
			}
			textWithUsers += fmt.Sprintf("%s %s %s ID = %d\n", users[i].FirstName, users[i].LastName, userName, users[i].ID)
		}

		bot.SendText(userID, textWithUsers)
		bot.SendText(userID, "Send back the ID of the user")
	}

	return shouldContinue
}

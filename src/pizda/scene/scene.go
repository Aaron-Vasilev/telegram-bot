package scene

import (
	"bot/src/bot"
	"bot/src/pizda/db"
	"bot/src/pizda/utils"
	common "bot/src/utils"
	t "bot/src/utils/types"
	"fmt"
)

func RegisterScenes(b *bot.Bot) {
	b.RegisterScene(utils.ForwardAll, ForwardAll)
}

func ForwardAll(bot *bot.Bot, u t.Update) {
	userID, _ := common.UserIdFromUpdate(u)
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
			bot.SendText(userID, common.WrongMsg)
			return
		}

		state.Data = u.Message.MessageID
		bot.SetCtxValue(userID, state)

		bot.SendMessage(t.Message{
			ChatId:      userID,
			Text:        "Are you sure you want to send this message?",
			ReplyMarkup: &common.ConformationInlineKeyboard,
		})
	case 3:
		if u.CallbackQuery == nil {
			bot.SendText(userID, common.WrongMsg)
			bot.EndCtx(userID)
			return
		}

		state, ok := bot.GetCtxValue(userID)
		if !ok {
			bot.SendText(userID, common.WrongMsg)
			bot.EndCtx(userID)
			return
		}

		messageID, ok := state.Data.(int)
		if !ok {
			bot.SendText(userID, common.WrongMsg)
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

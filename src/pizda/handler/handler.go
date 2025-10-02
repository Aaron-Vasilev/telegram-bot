package handler

import (
	"bot/src/bot"
	"bot/src/pizda/commands"
	"bot/src/pizda/db"
	"bot/src/utils"
	t "bot/src/utils/types"
	"strings"
)

func handleCallbackQuery(bot *bot.Bot, u t.Update) {
	data := u.CallbackData()

	if string(db.PizdaPaymentMethodBIT) == data {
	}
}

func HandleUpdate(bot *bot.Bot, u t.Update) {
	if u.Message == nil || strings.HasPrefix(u.Message.Text, "/") {
		handleMenu(bot, u)

		return
	}

	_, updateWithCallbackQuery := utils.UserIdFromUpdate(u)

	if updateWithCallbackQuery {
		handleCallbackQuery(bot, u)
	}
}

func HandleUpdates(bot *bot.Bot, updates []t.Update) {
	for _, update := range updates {
		HandleUpdate(bot, update)

		bot.Offset = update.UpdateID + 1
	}
}

func handleMenu(bot *bot.Bot, u t.Update) {
	commands.Start(bot, u)
}

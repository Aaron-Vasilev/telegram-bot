package main

import (
	"bot/src/bot"
	"bot/src/cron"
	"bot/src/db"
	"bot/src/handler"
	"bot/src/utils"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	utils.LoadEnv()

	bot := bot.NewBot(os.Getenv("TOKEN"))
	connection := db.ConnectDB(bot)
	defer connection.Close(bot.Ctx)
	cron.Cron(bot)

	if bot.IsProd {
		bot.StartWebhook(handler.HandleUpdate)
	} else {
		bot.StartLongPulling(handler.HandleUpdates)
	}
}
//TODO FAILD if user sends sticker


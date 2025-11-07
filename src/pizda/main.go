package main

import (
	"bot/src/bot"
	"bot/src/pizda/db"
	"bot/src/pizda/handler"
	"bot/src/utils"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	utils.LoadEnv("./src/pizda/.env")

	bot := bot.NewBot(os.Getenv("TOKEN"))
	connection := db.ConnectDB(bot)
	defer connection.Close(bot.Ctx)

	fmt.Println("Started!")
	if bot.IsProd {
		bot.StartWebhook(handler.HandleUpdate)
	} else {
		bot.StartLongPulling(handler.HandleUpdates)
	}
}

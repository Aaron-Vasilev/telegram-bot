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
	utils.LoadEnv()

	bot := bot.NewBot(os.Getenv("TOKEN"))
	connection := db.ConnectDB(bot)
	defer connection.Close(bot.Ctx)

	fmt.Println("Started!")
	bot.StartLongPulling(handler.HandleUpdates)
	fmt.Println("Finished")
}

package main

import (
	"bot/src/bot"
	"bot/src/cron"
	"bot/src/db"
	"bot/src/handler"
	onlinedb "bot/src/online/db"
	"bot/src/payment"
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
	cron.Cron(bot)

	payment.StartPaymentServer(bot)
	payment.StartSubscriptionServer(bot)

	fmt.Println("Started")
	if bot.IsProd {
		bot.StartWebhook(handler.HandleUpdate)
	} else {
		onlineConn := onlinedb.ConnectDB(bot)
		defer onlineConn.Close(bot.Ctx)

		bot.StartHTTPServer()
		bot.StartLongPulling(handler.HandleUpdates)
	}
}

//TODO FAILD if user sends sticker

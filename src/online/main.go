package main

import (
	"bot/src/bot"
	"bot/src/online/cron"
	"bot/src/online/db"
	"bot/src/online/handler"
	"bot/src/online/payment"
	"bot/src/online/scene"
	"bot/src/utils"
	"fmt"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	utils.LoadEnv()

	bot := bot.NewBot(os.Getenv("TOKEN"))
	connection := db.ConnectDB(bot)
	defer connection.Close(bot.Ctx)
	scene.RegisterScenes(bot)
	cron.Cron(bot)
	payment.StartPaymentServer(bot)

	bot.Mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	fmt.Println("Started!")
	if bot.IsProd {
		bot.StartWebhook(handler.HandleUpdate)
	} else {
		bot.StartHTTPServer()
		bot.StartLongPulling(handler.HandleUpdates)
	}
}

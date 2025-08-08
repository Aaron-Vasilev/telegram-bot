package main

import (
	"bot/src/bot"
	"bot/src/cron"
	"bot/src/db"
	"bot/src/handler"
	"bot/src/utils"
	"context"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

func NewBot(token string) *bot.Bot {
	ctx := context.Background()
	defer ctx.Done()

	return &bot.Bot{
		Token:   token,
		Offset:  0,
		IsDebug: os.Getenv("ENV") == "DEBUG",
		Ctx:     ctx,
	}
}

func main() {
	utils.LoadEnv()

	bot := NewBot(os.Getenv("TOKEN"))
	connection := db.ConnectDB(bot)
	defer connection.Close(bot.Ctx)
	cron.Cron(bot)

	fmt.Println("Launch!")

	for {
		updates := bot.GetUpdates()

		handler.HandleUpdates(bot, updates)
	}
}

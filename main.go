package main

import (
	"bot/src/bot"
	"bot/src/handler"
	"bot/src/utils"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"database/sql"

	_ "github.com/lib/pq"
)

func NewBot(token string) *bot.Bot {
	return &bot.Bot{
		Token: token,
		Offset: 0,
	}
}

func main() {
	utils.LoadEnv()

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))

	if err = db.Ping(); err != nil {
		log.Fatal("Error connecting to db", err)
	}
	defer db.Close()

	bot := NewBot(os.Getenv("TOKEN"))

	bot.IsDebug = os.Getenv("ENV") == "DEBUG"

	ctx := context.Background()
	defer ctx.Done()

	fmt.Println("Launch!")
	for {
		updates := bot.GetUpdates()

		ctx = handler.HandleUpdates(ctx, bot, db, updates)
		time.Sleep(3000 * time.Millisecond)
	}
}

package main

import (
	"bot/src/actions"
	"bot/src/bot"
	"bot/src/utils"
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

	fmt.Println("Launch!")
	for {
		updates := bot.GetUpdates()

		commands.HandleUpdates(bot, db, updates)
		time.Sleep(3000 * time.Millisecond)
	}
}

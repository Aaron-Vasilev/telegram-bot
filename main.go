package main

import (
	"bot/src/bot"
	"bot/src/handler"
	"bot/src/scene"
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
		Token:  token,
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
	ctx := scene.NewSceneContext()

	ctx.SetValue(111, scene.SceneState{
		Scene: "lollul",
		Stage: 0,
	})

	fmt.Println("Launch!")
	defer func() {
		if r := recover(); r != nil {
			var r interface{} = "this is a string"

			bot.Error("Panic! " + r.(string))
		}
	}()
	for {
		updates := bot.GetUpdates()

		handler.HandleUpdates(&ctx, bot, db, updates)
		time.Sleep(3000 * time.Millisecond)
	}
}

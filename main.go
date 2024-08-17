package main

import (
	"bot/src/actions/commands"
	"bot/src/utils"
	"fmt"
	"os"
	"time"
)

func NewBot(token string) *commands.Bot {
	return &commands.Bot{
		Token: token,
		Offset: 0,
	}
}

func main() {
	utils.LoadEnv()
	bot := NewBot(os.Getenv("TOKEN"))

	bot.IsDebug = os.Getenv("ENV") == "DEBUG"

	fmt.Println("Launch!")
	for {
		updates := bot.GetUpdates()

		commands.HandleUpdates(bot, updates)
		time.Sleep(3000 * time.Millisecond)
	}
}

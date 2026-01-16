package cron

import (
	"bot/src/bot"
	"bot/src/pizda/action"
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

func Cron(bot *bot.Bot) {
	location, err := time.LoadLocation("Asia/Jerusalem")

	if err != nil {
		log.Fatal("Error loading time zone:", err)
	}

	c := cron.New(cron.WithLocation(location))

	c.AddFunc("0 9,12,15,18,21 * * *", func() {
		action.NotifyAboutPaymentExpiration(bot)
	})

	go c.Start()
}

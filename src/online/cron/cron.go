package cron

import (
	"bot/src/bot"
	"bot/src/online/action"
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

	c.AddFunc("0 10 * * *", func() {
		action.NotifyAboutSubscriptionEnds(bot)
	})

	c.AddFunc("0 11 * * *", func() {
		action.KickExpiredUsers(bot)
	})

	go c.Start()
}

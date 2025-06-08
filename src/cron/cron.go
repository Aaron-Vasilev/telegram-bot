package cron

import (
	"bot/src/action"
	"bot/src/bot"
	"database/sql"
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

// * * * * *
// | | | | |
// | | | | |__ Day of the week (0 - 6) (Sunday to Saturday)
// | | | |____ Month (1 - 12)
// | | |______ Day of the month (1 - 31)
// | |________ Hour (0 - 23)
// |__________ Minute (0 - 59)

func Cron(bot *bot.Bot, db *sql.DB) {
	location, err := time.LoadLocation("Asia/Jerusalem")

	if err != nil {
		log.Fatal("Error loading time zone:", err)
	}

	c := cron.New(cron.WithLocation(location))

	c.AddFunc("0 10 * * 0", func() {
		action.NotifyAboutSubscriptionEnds(bot, db)
	})

	c.AddFunc("* 23 * * *", func() {
		action.NotifyAboutTommorowLesson(bot, db)
	})

	go c.Start()
}

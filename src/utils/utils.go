package utils

import (
	t "bot/src/utils/types"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func LoadEnv() {
  env := os.Getenv("ENV")

  if env == "" { 
    err := godotenv.Load()

    if err != nil {
      log.Fatal("No .env")
    }
  }
}

func GenerateTimetable(lessons []t.Lesson, showId bool) t.Message {
	var buttons [][]t.InlineKeyboardButton

	for _, l := range lessons {
		weekday := l.Date.Format("Mon")
		date := l.Date.Format("02/01")
		time := l.Time.Format("15:04")

		label := fmt.Sprintf("%s 🌀 %s 🌀 %s", weekday, date, time)

		if showId {
			label += ` ID = ${lesson.id}`
		} 

		var button []t.InlineKeyboardButton

		button = append(button, t.InlineKeyboardButton{
			Text: label,
			CallbackData: label,
		})
		buttons = append(buttons, button)
	}

	if len(lessons) == 0 {
		return t.Message{
			Text: "The timetable is not ready yet",
		}
	} else {
		return t.Message{
			Text: "🗓 Choose a day:",
			ReplyMarkup: t.InlineKeyboardMarkup{
				InlineKeyboard: buttons,
			},
		}
	}
}

func BeautyDate(date time.Time) string {
	return date.Format("02/01")
}

func GenerateKeyboard() *t.ReplyKeyboardMarkup {
	replyKeyboard := t.ReplyKeyboardMarkup{
		Keyboard: [][]t.KeyboardButton{ 
			{
				{
					Text: Keyboard["Timetable 🗓"],
				},
				{
					Text: Keyboard["Leaderboard 🏆"],
				},
			},
			{
				{
					Text: Keyboard["Profile 🧘"],
				},
				{
					Text: Keyboard["Contact 💌"],
				},
			},
		},
		ResizeKeyboard: true,
	}

	return &replyKeyboard
}

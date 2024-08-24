package utils

import (
	t "bot/src/utils/types"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
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
			CallbackData: fmt.Sprintf("SHOW_LESSON=%d", l.Id),
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

func GenerateLessonMessage(lessons []t.LessonWithUsers, userId int64) t.Message {
	var msg t.Message
	var buttons [][]t.InlineKeyboardButton
	var button []t.InlineKeyboardButton
	weekday := lessons[0].Date.Format("Mon")
	date := lessons[0].Date.Format("02/01")
	time := lessons[0].Time.Format("15:04")
	description := lessons[0].Description
	isUserInLesson := false

	for _, l := range lessons {
		if l.UserId != nil && *l.UserId == userId {
			isUserInLesson = true
			break
		}
	}

	if isUserInLesson {
		button = append(button, t.InlineKeyboardButton{
			Text: "Unregister from the lesson",
			CallbackData: fmt.Sprintf("UNREGISTER=%d", lessons[0].LessonId),
		})
	} else {
		button = append(button, t.InlineKeyboardButton{
			Text: "Register for the lesson",
			CallbackData: fmt.Sprintf("REGISTER=%d", lessons[0].LessonId),
		})
	}
	buttons = append(buttons, button)

	msg.ChatId = userId
	msg.Text = fmt.Sprintf(
		"%s 🚀 %s 🚀 %s 🕒\n%s\n\n%s",
		weekday, date, time, description, yogis(lessons),
		)
	msg.ParseMode = "html"
	msg.ReplyMarkup = t.InlineKeyboardMarkup{
		InlineKeyboard: buttons,
	}

	return msg
}

func yogis(lessons []t.LessonWithUsers) string {
	registered := 0
	students := ""

	if lessons[0].UserId != nil {
		for i, l := range lessons {
			name := *l.Name

			if l.Username != nil {
				name = "@" + *l.Username
			}

			students += fmt.Sprintf("\n%d. %s", i + 1, name)
			registered++
		}
	}

	yogs := fmt.Sprintf("Booked: <b>%d</b>/%d", registered, lessons[0].Max)

	return yogs + students
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

func IsAdmin(id int64) bool {
	adminsStr := os.Getenv("ADMIN")

	admins := strings.Split(adminsStr, ",")

	for _, adminStr := range(admins) {

		admin, err := strconv.ParseInt(adminStr, 10, 64)

		if err == nil && admin == id {
			return true
		}
	}

	return false
}

func UserIdFromUpdate(u t.Update) (int64, bool) {
	var userId int64
	updateWithCallbackQuery := u.CallbackQuery != nil

	if updateWithCallbackQuery {
		userId = u.CallbackQuery.From.ID
	} else if u.Message != nil {
		userId = u.Message.From.ID
	}

	return userId, updateWithCallbackQuery 
}


func UpdateMembership(memb *t.Membership, token t.Token) {
	if memb.Ends.After(token.Created) {
		membershipLasts := 28

		memb.Ends = memb.Ends.AddDate(0, 0, membershipLasts)
	} else {
		daysRemaining := 27

		memb.LessonsAvailable = 0
		memb.Starts = token.Created
		memb.Ends = token.Created.AddDate(0, 0, daysRemaining)
	}

	if token.Type == 1 {
		memb.LessonsAvailable += 4 
	} else if token.Type == 2 {
		memb.LessonsAvailable += 8 
	} else if token.Type == 8 {
		memb.LessonsAvailable = 0
	}
	memb.Type = token.Type
}

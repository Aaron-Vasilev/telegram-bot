package utils

import (
	t "bot/src/utils/types"
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

var (
	strongEmojis = []string{"😮‍💨", "🥳", "😎", "😰", "😤", "💯", "💦", "🖐️", "👌", "✊", "👊", "🙌", "🤳", "🦾", "🦵", "👂", "👃", "🫀", "🫁", "🏃‍♂️", "🏃‍♀️", "🧗‍♂️", "🧗‍♀️", "🏋️‍♂️", "🏋️‍♀️", "🤼‍♂️", "🤼‍♀️", "🦍", "🐜", "🥗", "🔥", "✨", "🎉", "🎊", "🥊", "🎽", "✅", "🙏", "☯️", "🧘🏿‍♀️", "🧘🏿‍♂️"}
	weakEmojis   = []string{"🤕", "🥀", "𓀗", "🥴", "🤒", "💀", "💔", "🥺", "👎🏻", "😔", "🫠", "🗿", "💩", "🐌", "🩸", "🍆🍑😩👉👌💦", "👶🏿", "🌡️", "💔😢😭", "🤮", "👴🏻", "🦠", "🥶", "🚽", "🙅🏿‍♀️", "🪨"}
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

func GenerateTimetableMsg(lessons []t.Lesson, showId bool) t.Message {
	var buttons [][]t.InlineKeyboardButton

	for _, l := range lessons {
		weekday := l.Date.Format("Mon")
		date := l.Date.Format("02/01")
		time := l.Time.Format("15:04")

		label := fmt.Sprintf("%s 🌀 %s 🌀 %s", weekday, date, time)

		if showId {
			label += fmt.Sprintf(" ID = %d", l.ID)
		}

		var button []t.InlineKeyboardButton

		button = append(button, t.InlineKeyboardButton{
			Text:         label,
			CallbackData: fmt.Sprintf("LESSON=%d", l.ID),
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

func GenerateLessonMessage(lesson t.LessonWithUsers, userId int64) t.Message {
	var msg t.Message
	var buttons [][]t.InlineKeyboardButton
	var button []t.InlineKeyboardButton
	weekday := lesson.Date.Format("Mon")
	date := lesson.Date.Format("02/01")
	time := lesson.Time.Format("15:04")
	description := lesson.Description
	isUserInLesson := false

	for _, u := range lesson.Users {
		if u.ID == userId {
			isUserInLesson = true
			break
		}
	}

	if isUserInLesson {
		button = append(button, t.InlineKeyboardButton{
			Text:         "Unregister from the lesson",
			CallbackData: fmt.Sprintf("%s=%d", UNREGISTER, lesson.LessonId),
		})
	} else {
		button = append(button, t.InlineKeyboardButton{
			Text:         "Register for the lesson",
			CallbackData: fmt.Sprintf("%s=%d", REGISTER, lesson.LessonId),
		})
	}
	buttons = append(buttons, button)

	msg.ChatId = userId
	msg.Text = fmt.Sprintf(
		"%s 🚀 %s 🚀 %s 🕒\n%s\n\n%s",
		weekday, date, time, description, yogis(lesson),
	)
	msg.ParseMode = "html"
	msg.ReplyMarkup = t.InlineKeyboardMarkup{
		InlineKeyboard: buttons,
	}

	return msg
}

func yogis(lesson t.LessonWithUsers) string {
	registered := 0
	students := ""

	for i, l := range lesson.Users {
		name := l.Name

		if l.Username != "" {
			name = "@" + l.Username
		}

		students += fmt.Sprintf("\n%d. %s %s", i+1, name, l.Emoji)
		registered++
	}

	yogs := fmt.Sprintf("Booked: <b>%d</b>/%d", registered, lesson.Max)

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

	for _, adminStr := range admins {

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

func ProfileText(u t.UserMembership) string {
	text := fmt.Sprintf("Your emoji:\n%s%s%s\n\nMembership ends:\n", u.User.Emoji, u.User.Emoji, u.User.Emoji)

	if u.Type == nil {
		text += "Sweetie🍪, you don't have one"
	} else {
		ends := u.Ends.Format("2006-01-02")
		text += fmt.Sprintf("<b>%s</b>\n\n", ends)

		if *u.Type == NoLimit {
			text += "<b>You</b> are <b>my</b> favourite student🤍"
		} else {
			text += fmt.Sprintf("Lessons remaining:\n <b>%d</b>", *u.LessonsAvailable)
		}
	}

	return text
}

func IsEmoji(s string) bool {
	emojiRe := regexp.MustCompile(`^([#*0-9]\x{FE0F}?\x{20E3}|©\x{FE0F}?|[®\x{203C}\x{2049}\x{2122}\x{2139}\x{2194}-\x{2199}\x{21A9}\x{21AA}]\x{FE0F}?|[\x{231A}\x{231B}]|[\x{2328}\x{23CF}]\x{FE0F}?|[\x{23E9}-\x{23EC}]|[\x{23ED}-\x{23EF}]\x{FE0F}?|\x{23F0}|[\x{23F1}\x{23F2}]\x{FE0F}?|\x{23F3}|[\x{23F8}-\x{23FA}\x{24C2}\x{25AA}\x{25AB}\x{25B6}\x{25C0}\x{25FB}\x{25FC}]\x{FE0F}?|[\x{25FD}\x{25FE}]|[\x{2600}-\x{2604}\x{260E}\x{2611}]\x{FE0F}?|[\x{2614}\x{2615}]|\x{2618}\x{FE0F}?|\x{261D}[\x{FE0F}\x{1F3FB}-\x{1F3FF}]?|[\x{2620}\x{2622}\x{2623}\x{2626}\x{262A}\x{262E}\x{262F}\x{2638}-\x{263A}\x{2640}\x{2642}]\x{FE0F}?|[\x{2648}-\x{2653}]|[\x{265F}\x{2660}\x{2663}\x{2665}\x{2666}\x{2668}\x{267B}\x{267E}]\x{FE0F}?|\x{267F}|\x{2692}\x{FE0F}?|\x{2693}|[\x{2694}-\x{2697}\x{2699}\x{269B}\x{269C}\x{26A0}]\x{FE0F}?|\x{26A1}|\x{26A7}\x{FE0F}?|[\x{26AA}\x{26AB}]|[\x{26B0}\x{26B1}]\x{FE0F}?|[\x{26BD}\x{26BE}\x{26C4}\x{26C5}]|\x{26C8}\x{FE0F}?|\x{26CE}|[\x{26CF}\x{26D1}\x{26D3}]\x{FE0F}?|\x{26D4}|\x{26E9}\x{FE0F}?|\x{26EA}|[\x{26F0}\x{26F1}]\x{FE0F}?|[\x{26F2}\x{26F3}]|\x{26F4}\x{FE0F}?|\x{26F5}|[\x{26F7}\x{26F8}]\x{FE0F}?|\x{26F9}(?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{FE0F}\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|[\x{26FA}\x{26FD}]|\x{2702}\x{FE0F}?|\x{2705}|[\x{2708}\x{2709}]\x{FE0F}?|[\x{270A}\x{270B}][\x{1F3FB}-\x{1F3FF}]?|[\x{270C}\x{270D}][\x{FE0F}\x{1F3FB}-\x{1F3FF}]?|\x{270F}\x{FE0F}?|[\x{2712}\x{2714}\x{2716}\x{271D}\x{2721}]\x{FE0F}?|\x{2728}|[\x{2733}\x{2734}\x{2744}\x{2747}]\x{FE0F}?|[\x{274C}\x{274E}\x{2753}-\x{2755}\x{2757}]|\x{2763}\x{FE0F}?|\x{2764}(?:\x{200D}[\x{1F525}\x{1FA79}]|\x{FE0F}(?:\x{200D}[\x{1F525}\x{1FA79}])?)?|[\x{2795}-\x{2797}]|\x{27A1}\x{FE0F}?|[\x{27B0}\x{27BF}]|[\x{2934}\x{2935}\x{2B05}-\x{2B07}]\x{FE0F}?|[\x{2B1B}\x{2B1C}\x{2B50}\x{2B55}]|[\x{3030}\x{303D}\x{3297}\x{3299}]\x{FE0F}?|[\x{1F004}\x{1F0CF}]|[\x{1F170}\x{1F171}\x{1F17E}\x{1F17F}]\x{FE0F}?|[\x{1F18E}\x{1F191}-\x{1F19A}]|\x{1F1E6}[\x{1F1E8}-\x{1F1EC}\x{1F1EE}\x{1F1F1}\x{1F1F2}\x{1F1F4}\x{1F1F6}-\x{1F1FA}\x{1F1FC}\x{1F1FD}\x{1F1FF}]|\x{1F1E7}[\x{1F1E6}\x{1F1E7}\x{1F1E9}-\x{1F1EF}\x{1F1F1}-\x{1F1F4}\x{1F1F6}-\x{1F1F9}\x{1F1FB}\x{1F1FC}\x{1F1FE}\x{1F1FF}]|\x{1F1E8}[\x{1F1E6}\x{1F1E8}\x{1F1E9}\x{1F1EB}-\x{1F1EE}\x{1F1F0}-\x{1F1F5}\x{1F1F7}\x{1F1FA}-\x{1F1FF}]|\x{1F1E9}[\x{1F1EA}\x{1F1EC}\x{1F1EF}\x{1F1F0}\x{1F1F2}\x{1F1F4}\x{1F1FF}]|\x{1F1EA}[\x{1F1E6}\x{1F1E8}\x{1F1EA}\x{1F1EC}\x{1F1ED}\x{1F1F7}-\x{1F1FA}]|\x{1F1EB}[\x{1F1EE}-\x{1F1F0}\x{1F1F2}\x{1F1F4}\x{1F1F7}]|\x{1F1EC}[\x{1F1E6}\x{1F1E7}\x{1F1E9}-\x{1F1EE}\x{1F1F1}-\x{1F1F3}\x{1F1F5}-\x{1F1FA}\x{1F1FC}\x{1F1FE}]|\x{1F1ED}[\x{1F1F0}\x{1F1F2}\x{1F1F3}\x{1F1F7}\x{1F1F9}\x{1F1FA}]|\x{1F1EE}[\x{1F1E8}-\x{1F1EA}\x{1F1F1}-\x{1F1F4}\x{1F1F6}-\x{1F1F9}]|\x{1F1EF}[\x{1F1EA}\x{1F1F2}\x{1F1F4}\x{1F1F5}]|\x{1F1F0}[\x{1F1EA}\x{1F1EC}-\x{1F1EE}\x{1F1F2}\x{1F1F3}\x{1F1F5}\x{1F1F7}\x{1F1FC}\x{1F1FE}\x{1F1FF}]|\x{1F1F1}[\x{1F1E6}-\x{1F1E8}\x{1F1EE}\x{1F1F0}\x{1F1F7}-\x{1F1FB}\x{1F1FE}]|\x{1F1F2}[\x{1F1E6}\x{1F1E8}-\x{1F1ED}\x{1F1F0}-\x{1F1FF}]|\x{1F1F3}[\x{1F1E6}\x{1F1E8}\x{1F1EA}-\x{1F1EC}\x{1F1EE}\x{1F1F1}\x{1F1F4}\x{1F1F5}\x{1F1F7}\x{1F1FA}\x{1F1FF}]|\x{1F1F4}\x{1F1F2}|\x{1F1F5}[\x{1F1E6}\x{1F1EA}-\x{1F1ED}\x{1F1F0}-\x{1F1F3}\x{1F1F7}-\x{1F1F9}\x{1F1FC}\x{1F1FE}]|\x{1F1F6}\x{1F1E6}|\x{1F1F7}[\x{1F1EA}\x{1F1F4}\x{1F1F8}\x{1F1FA}\x{1F1FC}]|\x{1F1F8}[\x{1F1E6}-\x{1F1EA}\x{1F1EC}-\x{1F1F4}\x{1F1F7}-\x{1F1F9}\x{1F1FB}\x{1F1FD}-\x{1F1FF}]|\x{1F1F9}[\x{1F1E6}\x{1F1E8}\x{1F1E9}\x{1F1EB}-\x{1F1ED}\x{1F1EF}-\x{1F1F4}\x{1F1F7}\x{1F1F9}\x{1F1FB}\x{1F1FC}\x{1F1FF}]|\x{1F1FA}[\x{1F1E6}\x{1F1EC}\x{1F1F2}\x{1F1F3}\x{1F1F8}\x{1F1FE}\x{1F1FF}]|\x{1F1FB}[\x{1F1E6}\x{1F1E8}\x{1F1EA}\x{1F1EC}\x{1F1EE}\x{1F1F3}\x{1F1FA}]|\x{1F1FC}[\x{1F1EB}\x{1F1F8}]|\x{1F1FD}\x{1F1F0}|\x{1F1FE}[\x{1F1EA}\x{1F1F9}]|\x{1F1FF}[\x{1F1E6}\x{1F1F2}\x{1F1FC}]|\x{1F201}|\x{1F202}\x{FE0F}?|[\x{1F21A}\x{1F22F}\x{1F232}-\x{1F236}]|\x{1F237}\x{FE0F}?|[\x{1F238}-\x{1F23A}\x{1F250}\x{1F251}\x{1F300}-\x{1F320}]|[\x{1F321}\x{1F324}-\x{1F32C}]\x{FE0F}?|[\x{1F32D}-\x{1F335}]|\x{1F336}\x{FE0F}?|[\x{1F337}-\x{1F37C}]|\x{1F37D}\x{FE0F}?|[\x{1F37E}-\x{1F384}]|\x{1F385}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F386}-\x{1F393}]|[\x{1F396}\x{1F397}\x{1F399}-\x{1F39B}\x{1F39E}\x{1F39F}]\x{FE0F}?|[\x{1F3A0}-\x{1F3C1}]|\x{1F3C2}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F3C3}\x{1F3C4}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|[\x{1F3C5}\x{1F3C6}]|\x{1F3C7}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F3C8}\x{1F3C9}]|\x{1F3CA}(?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|[\x{1F3CB}\x{1F3CC}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{FE0F}\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|[\x{1F3CD}\x{1F3CE}]\x{FE0F}?|[\x{1F3CF}-\x{1F3D3}]|[\x{1F3D4}-\x{1F3DF}]\x{FE0F}?|[\x{1F3E0}-\x{1F3F0}]|\x{1F3F3}(?:\x{200D}(?:\x{26A7}\x{FE0F}?|\x{1F308})|\x{FE0F}(?:\x{200D}(?:\x{26A7}\x{FE0F}?|\x{1F308}))?)?|\x{1F3F4}(?:\x{200D}\x{2620}\x{FE0F}?|\x{E0067}\x{E0062}(?:\x{E0065}\x{E006E}\x{E0067}|\x{E0073}\x{E0063}\x{E0074}|\x{E0077}\x{E006C}\x{E0073})\x{E007F})?|[\x{1F3F5}\x{1F3F7}]\x{FE0F}?|[\x{1F3F8}-\x{1F407}]|\x{1F408}(?:\x{200D}\x{2B1B})?|[\x{1F409}-\x{1F414}]|\x{1F415}(?:\x{200D}\x{1F9BA})?|[\x{1F416}-\x{1F43A}]|\x{1F43B}(?:\x{200D}\x{2744}\x{FE0F}?)?|[\x{1F43C}-\x{1F43E}]|\x{1F43F}\x{FE0F}?|\x{1F440}|\x{1F441}(?:\x{200D}\x{1F5E8}\x{FE0F}?|\x{FE0F}(?:\x{200D}\x{1F5E8}\x{FE0F}?)?)?|[\x{1F442}\x{1F443}][\x{1F3FB}-\x{1F3FF}]?|[\x{1F444}\x{1F445}]|[\x{1F446}-\x{1F450}][\x{1F3FB}-\x{1F3FF}]?|[\x{1F451}-\x{1F465}]|[\x{1F466}\x{1F467}][\x{1F3FB}-\x{1F3FF}]?|\x{1F468}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:\x{1F48B}\x{200D})?\x{1F468}|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}]|\x{1F466}(?:\x{200D}\x{1F466})?|\x{1F467}(?:\x{200D}[\x{1F466}\x{1F467}])?|[\x{1F468}\x{1F469}]\x{200D}(?:\x{1F466}(?:\x{200D}\x{1F466})?|\x{1F467}(?:\x{200D}[\x{1F466}\x{1F467}])?)|[\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}])|\x{1F3FB}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:\x{1F48B}\x{200D})?\x{1F468}[\x{1F3FB}-\x{1F3FF}]|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}\x{1F468}[\x{1F3FC}-\x{1F3FF}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?|\x{1F3FC}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:\x{1F48B}\x{200D})?\x{1F468}[\x{1F3FB}-\x{1F3FF}]|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}\x{1F468}[\x{1F3FB}\x{1F3FD}-\x{1F3FF}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?|\x{1F3FD}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:\x{1F48B}\x{200D})?\x{1F468}[\x{1F3FB}-\x{1F3FF}]|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}\x{1F468}[\x{1F3FB}\x{1F3FC}\x{1F3FE}\x{1F3FF}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?|\x{1F3FE}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:\x{1F48B}\x{200D})?\x{1F468}[\x{1F3FB}-\x{1F3FF}]|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}\x{1F468}[\x{1F3FB}-\x{1F3FD}\x{1F3FF}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?|\x{1F3FF}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:\x{1F48B}\x{200D})?\x{1F468}[\x{1F3FB}-\x{1F3FF}]|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}\x{1F468}[\x{1F3FB}-\x{1F3FE}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?)?|\x{1F469}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:\x{1F48B}\x{200D})?[\x{1F468}\x{1F469}]|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}]|\x{1F466}(?:\x{200D}\x{1F466})?|\x{1F467}(?:\x{200D}[\x{1F466}\x{1F467}])?|\x{1F469}\x{200D}(?:\x{1F466}(?:\x{200D}\x{1F466})?|\x{1F467}(?:\x{200D}[\x{1F466}\x{1F467}])?)|[\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}])|\x{1F3FB}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:[\x{1F468}\x{1F469}][\x{1F3FB}-\x{1F3FF}]|\x{1F48B}\x{200D}[\x{1F468}\x{1F469}][\x{1F3FB}-\x{1F3FF}])|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}[\x{1F468}\x{1F469}][\x{1F3FC}-\x{1F3FF}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?|\x{1F3FC}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:[\x{1F468}\x{1F469}][\x{1F3FB}-\x{1F3FF}]|\x{1F48B}\x{200D}[\x{1F468}\x{1F469}][\x{1F3FB}-\x{1F3FF}])|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}[\x{1F468}\x{1F469}][\x{1F3FB}\x{1F3FD}-\x{1F3FF}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?|\x{1F3FD}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:[\x{1F468}\x{1F469}][\x{1F3FB}-\x{1F3FF}]|\x{1F48B}\x{200D}[\x{1F468}\x{1F469}][\x{1F3FB}-\x{1F3FF}])|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}[\x{1F468}\x{1F469}][\x{1F3FB}\x{1F3FC}\x{1F3FE}\x{1F3FF}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?|\x{1F3FE}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:[\x{1F468}\x{1F469}][\x{1F3FB}-\x{1F3FF}]|\x{1F48B}\x{200D}[\x{1F468}\x{1F469}][\x{1F3FB}-\x{1F3FF}])|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}[\x{1F468}\x{1F469}][\x{1F3FB}-\x{1F3FD}\x{1F3FF}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?|\x{1F3FF}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:[\x{1F468}\x{1F469}][\x{1F3FB}-\x{1F3FF}]|\x{1F48B}\x{200D}[\x{1F468}\x{1F469}][\x{1F3FB}-\x{1F3FF}])|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}[\x{1F468}\x{1F469}][\x{1F3FB}-\x{1F3FE}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?)?|\x{1F46A}|[\x{1F46B}-\x{1F46D}][\x{1F3FB}-\x{1F3FF}]?|\x{1F46E}(?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|\x{1F46F}(?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?|[\x{1F470}\x{1F471}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|\x{1F472}[\x{1F3FB}-\x{1F3FF}]?|\x{1F473}(?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|[\x{1F474}-\x{1F476}][\x{1F3FB}-\x{1F3FF}]?|\x{1F477}(?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|\x{1F478}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F479}-\x{1F47B}]|\x{1F47C}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F47D}-\x{1F480}]|[\x{1F481}\x{1F482}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|\x{1F483}[\x{1F3FB}-\x{1F3FF}]?|\x{1F484}|\x{1F485}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F486}\x{1F487}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|[\x{1F488}-\x{1F48E}]|\x{1F48F}[\x{1F3FB}-\x{1F3FF}]?|\x{1F490}|\x{1F491}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F492}-\x{1F4A9}]|\x{1F4AA}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F4AB}-\x{1F4FC}]|\x{1F4FD}\x{FE0F}?|[\x{1F4FF}-\x{1F53D}]|[\x{1F549}\x{1F54A}]\x{FE0F}?|[\x{1F54B}-\x{1F54E}\x{1F550}-\x{1F567}]|[\x{1F56F}\x{1F570}\x{1F573}]\x{FE0F}?|\x{1F574}[\x{FE0F}\x{1F3FB}-\x{1F3FF}]?|\x{1F575}(?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{FE0F}\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|[\x{1F576}-\x{1F579}]\x{FE0F}?|\x{1F57A}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F587}\x{1F58A}-\x{1F58D}]\x{FE0F}?|\x{1F590}[\x{FE0F}\x{1F3FB}-\x{1F3FF}]?|[\x{1F595}\x{1F596}][\x{1F3FB}-\x{1F3FF}]?|\x{1F5A4}|[\x{1F5A5}\x{1F5A8}\x{1F5B1}\x{1F5B2}\x{1F5BC}\x{1F5C2}-\x{1F5C4}\x{1F5D1}-\x{1F5D3}\x{1F5DC}-\x{1F5DE}\x{1F5E1}\x{1F5E3}\x{1F5E8}\x{1F5EF}\x{1F5F3}\x{1F5FA}]\x{FE0F}?|[\x{1F5FB}-\x{1F62D}]|\x{1F62E}(?:\x{200D}\x{1F4A8})?|[\x{1F62F}-\x{1F634}]|\x{1F635}(?:\x{200D}\x{1F4AB})?|\x{1F636}(?:\x{200D}\x{1F32B}\x{FE0F}?)?|[\x{1F637}-\x{1F644}]|[\x{1F645}-\x{1F647}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|[\x{1F648}-\x{1F64A}]|\x{1F64B}(?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|\x{1F64C}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F64D}\x{1F64E}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|\x{1F64F}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F680}-\x{1F6A2}]|\x{1F6A3}(?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|[\x{1F6A4}-\x{1F6B3}]|[\x{1F6B4}-\x{1F6B6}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|[\x{1F6B7}-\x{1F6BF}]|\x{1F6C0}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F6C1}-\x{1F6C5}]|\x{1F6CB}\x{FE0F}?|\x{1F6CC}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F6CD}-\x{1F6CF}]\x{FE0F}?|[\x{1F6D0}-\x{1F6D2}\x{1F6D5}-\x{1F6D7}\x{1F6DD}-\x{1F6DF}]|[\x{1F6E0}-\x{1F6E5}\x{1F6E9}]\x{FE0F}?|[\x{1F6EB}\x{1F6EC}]|[\x{1F6F0}\x{1F6F3}]\x{FE0F}?|[\x{1F6F4}-\x{1F6FC}\x{1F7E0}-\x{1F7EB}\x{1F7F0}]|\x{1F90C}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F90D}\x{1F90E}]|\x{1F90F}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F910}-\x{1F917}]|[\x{1F918}-\x{1F91F}][\x{1F3FB}-\x{1F3FF}]?|[\x{1F920}-\x{1F925}]|\x{1F926}(?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|[\x{1F927}-\x{1F92F}]|[\x{1F930}-\x{1F934}][\x{1F3FB}-\x{1F3FF}]?|\x{1F935}(?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|\x{1F936}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F937}-\x{1F939}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|\x{1F93A}|\x{1F93C}(?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?|[\x{1F93D}\x{1F93E}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|[\x{1F93F}-\x{1F945}\x{1F947}-\x{1F976}]|\x{1F977}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F978}-\x{1F9B4}]|[\x{1F9B5}\x{1F9B6}][\x{1F3FB}-\x{1F3FF}]?|\x{1F9B7}|[\x{1F9B8}\x{1F9B9}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|\x{1F9BA}|\x{1F9BB}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F9BC}-\x{1F9CC}]|[\x{1F9CD}-\x{1F9CF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|\x{1F9D0}|\x{1F9D1}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F384}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}\x{1F9D1}|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}])|\x{1F3FB}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:\x{1F48B}\x{200D}|)\x{1F9D1}[\x{1F3FC}-\x{1F3FF}]|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F384}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}\x{1F9D1}[\x{1F3FB}-\x{1F3FF}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?|\x{1F3FC}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:\x{1F48B}\x{200D}|)\x{1F9D1}[\x{1F3FB}\x{1F3FD}-\x{1F3FF}]|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F384}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}\x{1F9D1}[\x{1F3FB}-\x{1F3FF}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?|\x{1F3FD}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:\x{1F48B}\x{200D}|)\x{1F9D1}[\x{1F3FB}\x{1F3FC}\x{1F3FE}\x{1F3FF}]|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F384}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}\x{1F9D1}[\x{1F3FB}-\x{1F3FF}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?|\x{1F3FE}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:\x{1F48B}\x{200D}|)\x{1F9D1}[\x{1F3FB}-\x{1F3FD}\x{1F3FF}]|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F384}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}\x{1F9D1}[\x{1F3FB}-\x{1F3FF}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?|\x{1F3FF}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:\x{1F48B}\x{200D}|)\x{1F9D1}[\x{1F3FB}-\x{1F3FE}]|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F384}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}\x{1F9D1}[\x{1F3FB}-\x{1F3FF}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?)?|[\x{1F9D2}\x{1F9D3}][\x{1F3FB}-\x{1F3FF}]?|\x{1F9D4}(?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|\x{1F9D5}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F9D6}-\x{1F9DD}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|[\x{1F9DE}\x{1F9DF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?|[\x{1F9E0}-\x{1F9FF}\x{1FA70}-\x{1FA74}\x{1FA78}-\x{1FA7C}\x{1FA80}-\x{1FA86}\x{1FA90}-\x{1FAAC}\x{1FAB0}-\x{1FABA}\x{1FAC0}-\x{1FAC2}]|[\x{1FAC3}-\x{1FAC5}][\x{1F3FB}-\x{1F3FF}]?|[\x{1FAD0}-\x{1FAD9}\x{1FAE0}-\x{1FAE7}]|\x{1FAF0}[\x{1F3FB}-\x{1F3FF}]?|\x{1FAF1}(?:\x{1F3FB}(?:\x{200D}\x{1FAF2}[\x{1F3FC}-\x{1F3FF}])?|\x{1F3FC}(?:\x{200D}\x{1FAF2}[\x{1F3FB}\x{1F3FD}-\x{1F3FF}])?|\x{1F3FD}(?:\x{200D}\x{1FAF2}[\x{1F3FB}\x{1F3FC}\x{1F3FE}\x{1F3FF}])?|\x{1F3FE}(?:\x{200D}\x{1FAF2}[\x{1F3FB}-\x{1F3FD}\x{1F3FF}])?|\x{1F3FF}(?:\x{200D}\x{1FAF2}[\x{1F3FB}-\x{1F3FE}])?)?|[\x{1FAF2}-\x{1FAF6}][\x{1F3FB}-\x{1F3FF}])$`)

	return emojiRe.MatchString(s)
}

type ValidatedLesson struct {
	IsValid     bool
	Date        string
	Time        string
	Description string
	Max         string
}

func ValidateLessonStr(s string) ValidatedLesson {
	var lesson ValidatedLesson

	splited := strings.Split(s, "\n")

	if len(splited) != 4 {
		return lesson
	}

	for i := range splited {
		splited[i] = strings.TrimSpace(splited[i])
	}

	if len(splited[2]) == 0 ||
		!DateRegexp().Match([]byte(splited[0])) ||
		!TimeRegexp().Match([]byte(splited[1])) ||
		!NumRegexp().Match([]byte(splited[3])) {
		return lesson
	}

	lesson.IsValid = true
	lesson.Date = splited[0]
	lesson.Time = splited[1]
	lesson.Description = splited[2]
	lesson.Max = splited[3]

	return lesson
}

func UserMemText(u t.UserMembership) string {
	userName := "null"

	if u.User.Username != "" {
		userName = u.User.Username
	}

	if u.Ends == nil {
		return fmt.Sprintf("Type <b>Y</b> or <b>N</b>:\n <b>%s - %s</b>\n <b>No membership</b>",
			u.User.Name, userName)

	} else {
		ends := u.Ends.Format("2006-01-02")

		return fmt.Sprintf("Type <b>Y</b> or <b>N</b>:\n <b>%s - %s</b>\n Ends: <b>%s</b>\nLessons <b>%d</b>",
			u.User.Name, userName, ends, *u.LessonsAvailable)
	}
}

var medals = []string{"🥇", "🥈", "🥉"}

func medalEmoji(i int) string {
	if i < len(medals) {
		return medals[i]
	}
	return "🎖"
}

func LeaderboardText(users []t.UserAttendance, userID int64) string {
	text := "The Yoggiest Yogi of the month📅:\n\n"
	medalI := 0
	userHasSeen := false

	for i := 0; i < len(users); i++ {
		if users[i].U.ID == userID {
			userHasSeen = true
		}
		if i > 0 && users[i].Count < users[i-1].Count {
			medalI++
		}

		if userHasSeen && i > 4 {
			break
		}

		text += fmt.Sprintf("%s %s %s (%d times)\n",
			medalEmoji(medalI),
			users[i].U.Name,
			users[i].U.Emoji,
			users[i].Count,
		)
	}

	return text
}

func BuildInlineKeyboard(vals []string) t.InlineKeyboardMarkup {
	var btns []t.InlineKeyboardButton

	for _, val := range vals {
		btns = append(btns, t.InlineKeyboardButton{
			Text:         val,
			CallbackData: val,
		})
	}

	return t.InlineKeyboardMarkup{
		InlineKeyboard: [][]t.InlineKeyboardButton{
			btns,
		},
	}
}

func LessonRegexp() *regexp.Regexp {
	return regexp.MustCompile(`^LESSON=\d+$`)
}

func DateRegexp() *regexp.Regexp {
	return regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
}

func TimeRegexp() *regexp.Regexp {
	return regexp.MustCompile(`^\d{2}:\d{2}$`)
}

func NumRegexp() *regexp.Regexp {
	return regexp.MustCompile(`^\d+$`)
}

func FullName(firstName, lastName string) string {
	return strings.Trim(fmt.Sprintf("%s %s", firstName, lastName), " ")
}

func IfUserComesMsg(userId int64, lesson t.Lesson) t.Message {
	num1 := rand.Intn(len(strongEmojis))
	num2 := rand.Intn(len(weakEmojis))

	return t.Message{
		ChatId:    userId,
		Text:      fmt.Sprintf("Hey puncake🥞, are you comming tommorow at <b>%s</b>?", lesson.Time.Format("15:04")),
		ParseMode: "html",
		ReplyMarkup: t.InlineKeyboardMarkup{
			InlineKeyboard: [][]t.InlineKeyboardButton{
				{
					{
						Text:         fmt.Sprintf("Yes %s", strongEmojis[num1]),
						CallbackData: fmt.Sprintf("IF_USER_COMES=%d=%s", lesson.ID, YES),
					},
					{
						Text:         fmt.Sprintf("No %s", weakEmojis[num2]),
						CallbackData: fmt.Sprintf("IF_USER_COMES=%d=%s", lesson.ID, NO),
					},
				},
			},
		},
	}
}

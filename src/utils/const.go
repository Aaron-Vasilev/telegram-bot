package utils

import (
	t "bot/src/utils/types"
)

const (
	// Keyboard
	Timetable   = "Timetable 🗓"
	Leaderboard = "Leaderboard 🏆"
	Profile     = "Profile 🧘"
	Contact     = "Contact 💌"

	// ADMIN keyboard
	SignStudents       = "Sign students ✍🏿"
	AddLessons         = "Add lessons 📚"
	AssignMembership   = "Assign a membership 🔑"
	NotifyAboutLessons = "Notify about lessons 💬"
	FreezeMembership   = "Freeze memberships ❄️"
	ForwardAll         = "Forward all 💌"
	EditLesson         = "Edit lesson ✏️"

	// Inline keyboard
	ChangeEmoji = "Change emoji"
	HowToFind   = "How to find?"

	// Messages
	GreetingMsg          = "Hello to all my dear yoga students!\nI hope you are feeling healthy and happy.\nI look forward to practice together. See you on the mat🤍"
	WrongMsg             = "Oops, something went wrong :c"
	ContactMsg           = "Address: <b>Tel-Aviv Jaffa, Yefet Street 22</b>\n\nTelephone: <b>0534257328</b> \n\nQuestions: @vialettochka\n\n<b>Prices:</b>\n 70₪ - first time\n 90₪ - visit without a pass\n<b>4 weeks membership:</b>\n 280₪ - 1 lesson in a week\n 400₪ - 2 lessons in a week"
	SendEmojiMsg         = "Send me a message with only <b>one emoji</b>\n\n*Unfortunately, Telegram doesn't support their cool emojis for bots"
	AddLessonMsg         = "Send me a message with current format:\n\n2024-10-01\n10:00\nMorning yoga 60 MIN\n10"
	SeeYouMsg            = "See you in the lesson✨"
	NotANumberMsg        = "It's not a number 🔪"
	NoClassesMsg         = "Hey puncake🥞, I won't be able to give classes for a couple days. So, <b>I update the end</b> date of your membership🤝\nThis is your current membership:"
	UpdatedMembershipMsg = "And this is the updated one. See you at the lesson, chocolatka🍫"
	GoodJob              = "Good job 🌝"

	// Membership types
	Onece   = 1
	Twice   = 2
	NoLimit = 8

	// Stickers
	PinkSheepMeditating = "CAACAgIAAxkBAAEi0oFklVEDLgLxyg23P1fyOASMuSO7SQACbgAD5KDOByc3KCA4N217LwQ"
)

var (
	Keyboard = map[string]string{
		Timetable:   Timetable,
		Leaderboard: Leaderboard,
		Profile:     Profile,
		Contact:     Contact,
	}
	AdminKeyboard = []string{
		AddLessons,
		AssignMembership,
		EditLesson,
		ForwardAll,
		FreezeMembership,
		NotifyAboutLessons,
		SignStudents,
	}

	// Inline keyboard
	ConformationInlineKeyboard = t.InlineKeyboardMarkup{
		InlineKeyboard: [][]t.InlineKeyboardButton{
			{
				{
					Text:         "Yes",
					CallbackData: "YES",
				},
				{
					Text:         "No",
					CallbackData: "NO",
				},
			},
		},
	}
)

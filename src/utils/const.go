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
	Course      = "Course 📚"
	Prices      = "Prices 🏷️"

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
	ContactMsg           = "Address: <b>Tel-Aviv Jaffa, Abed El Rauf El Bitar 6</b>\n\nTelephone: <b>0534257328</b> \n\nQuestions: @vialettochka"
	PricesMsg            = "<b>Price for a <i>four weeks</i> membership</b>\n 280₪ - One lesson per week\n 400₪ - Two lessons per week\n\n<b>One time entrance:</b>\n 70₪ - first time\n 90₪ - visit without a pass"
	SendEmojiMsg         = "Send me a message with only <b>one emoji</b>\n\n*Unfortunately, Telegram doesn't support their cool emojis for bots"
	AddLessonMsg         = "Send me a message with current format:\n\n2025-10-01\n10:00\nMorning yoga 60 MIN\n10"
	SeeYouMsg            = "See you in the lesson✨"
	NotANumberMsg        = "It's not a number 🔪"
	NoClassesMsg         = "Hey puncake🥞, I won't be able to give classes for a couple days. So, <b>I update the end</b> date of your membership🤝\nThis is your current membership:"
	UpdatedMembershipMsg = "And this is the updated one. See you at the lesson, chocolatka🍫"
	GoodJob              = "Good job 🌝"
	YouAreFree           = "You are free, fatass...🌚"

	// Membership types
	Onece   = 1
	Twice   = 2
	NoLimit = 8

	// Stickers
	PinkSheepMeditating = "CAACAgIAAxkBAAEi0oFklVEDLgLxyg23P1fyOASMuSO7SQACbgAD5KDOByc3KCA4N217LwQ"

	YES        = "YES"
	NO         = "NO"
	UNREGISTER = "UNREGISTER"
	REGISTER   = "REGISTER"

	VIOLETTA_ID = 833382946
	MY_ID       = 362575139
)

var (
	Keyboard = []string{
		Timetable,
		Leaderboard,
		Profile,
		Contact,
		Course,
		Prices,
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

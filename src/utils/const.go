package utils

const (
	// Keyboard
	Timetable = "Timetable 🗓"
	Leaderboard = "Leaderboard 🏆"
	Profile = "Profile 🧘"
	Contact = "Contact 💌"

	// ADMIN
	GenerateToken = "Generate a token 🔑"

	// Messages
    Greeting = "Hello to all my dear yoga students!\nI hope you are feeling healthy and happy.\nI look forward to practice together. See you on the mat🤍"
	Wrong = "Oops, something went wrong :c"
)

var (
	Keyboard = map[string]string{ 
		Timetable: Timetable, 
		Leaderboard: Leaderboard,
		Profile: Profile,
		Contact: Contact,
	}
	AdminKeyboard = map[string]string{ 
		GenerateToken: GenerateToken,
	}
) 


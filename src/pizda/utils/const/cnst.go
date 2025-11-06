package cnst

const (
	// No pay keyboard
	Whom         = "Кому подходит? 🤔"
	Programm     = "Программа 📋"
	Purchase     = "Преобрести курс 💳"
	TestTraining = "Пробная тренировка 🧘🏻‍♀️"

	// Pay keyboard
	AssignSubscription = "Ученик оплатил 💳"
)

var (
	SaleKeyboard = []string{
		Whom,
		Programm,
		TestTraining,
		Purchase,
	}
	AdminKeyboard = []string{
		AssignSubscription,
	}
)

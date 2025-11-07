package cnst

const (
	// No pay keyboard
	Whom         = "Кому подходит? 🤔"
	Programm     = "Программа 📋"
	Purchase     = "Преобрести курс 💳"
	TestTraining = "Пробная тренировка 🧘🏻‍♀️"
	Prices       = "Цены и тарифы 🏷️"

	// Pay keyboard
	AssignSubscription = "Ученик оплатил 💳"
)

var (
	SaleKeyboard = []string{
		Whom,
		Programm,
		TestTraining,
		Prices,
		Purchase,
	}
	AdminKeyboard = []string{
		AssignSubscription,
	}
)

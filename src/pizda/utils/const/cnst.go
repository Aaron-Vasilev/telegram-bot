package cnst

const (
	// No pay keyboard
	Whom         = "Кому подходит? 🤔"
	Programm     = "Программа 📋"
	Purchase     = "Преобрести курс 💳"
	TestTraining = "Пробная тренировка 🧘🏻‍♀️"
	Prices       = "Цены и тарифы 🏷️"

	// Pay keyboard
	Lessons      = "Уроки 📚"
	Subscription = "Подписка 🎟️"

	// Admin keyboard
	AssignSubscription = "Ученик оплатил 💳"
	ForwardAll         = "Переслать всем сообщение"
)

var (
	SaleKeyboard = []string{
		Whom,
		Programm,
		TestTraining,
		Prices,
		Purchase,
	}
	PayKeyboard = []string{
		Lessons,
		Subscription,
	}
	AdminKeyboard = []string{
		AssignSubscription,
		ForwardAll,
	}
)

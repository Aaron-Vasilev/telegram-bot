package cnst

const (
	// No pay keyboard
	Whom         = "Кому подходит? 🤔"
	Programm     = "Программа 📋"
	Purchase     = "Преобрести курс 💳"
	TestTraining = "Пробные тренировки 🧘🏻‍♀️"
	Prices       = "Цены и тарифы 🏷️"

	// Pay keyboard
	Lessons      = "Уроки 📚"
	Subscription = "Подписка 🎟️"

	// Admin keyboard
	AssignSubscription = "Ученик оплатил 💳"
	ForwardAll         = "Переслать всем сообщение"
	ExtendPayment      = "Продлить подписку 🔄"

	Video              = "VIDEO"
	ErrorMsg           = "Что-по пошло не так, уже чиним 🛠️"
	HowToExtendPayment = "Как продлить подписку 🔄"

	HormoneYoga = "Гармональная йога ⚤"
	TazDno      = "Мышцы тазового дна 🏋️‍♀️"
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
		ExtendPayment,
		ForwardAll,
	}
)

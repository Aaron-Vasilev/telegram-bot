package cnst

const (
	Subscription = "Подписка 🎟️"
	Purchase     = "Оформить подписку 💳"

	AssignSubscription = "Выдать подписку 🔑"

	ErrorMsg       = "Что-то пошло не так, уже чиним 🛠️"
	GreetingMsg    = "Добро пожаловать в онлайн йога-клуб 🧘‍♀️\n\nЗдесь ты получишь доступ к живым занятиям в группе. Оформи подписку, чтобы присоединиться."
	NoActiveSubMsg = "У тебя пока нет активной подписки."
)

var (
	Keyboard = []string{
		Subscription,
		Purchase,
	}
	AdminKeyboard = []string{
		AssignSubscription,
	}
)

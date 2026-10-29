package cnst

const SubscriptionDays = 30

const (
	Purchase     = "Оформить подписку 💳"
	Subscription = "О подписке 🎟️"

	AssignSubscription = "Выдать подписку 🔑"

	ErrorMsg       = "Что-то пошло не так, уже чиним 🛠️"
	GreetingMsg    = "Добро пожаловать в онлайн йога-клуб 🧘‍♀️\n\nЗдесь ты получишь доступ к живым занятиям в группе. Оформи подписку, чтобы присоединиться."
	NoActiveSubMsg = "У тебя пока нет активной подписки."
)

var (
	Keyboard = []string{
		Purchase,
		Subscription,
	}
	AdminKeyboard = []string{
		AssignSubscription,
	}
)

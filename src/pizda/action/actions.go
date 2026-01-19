package action

import (
	"bot/src/bot"
	"bot/src/pizda/db"
	cnst "bot/src/pizda/utils/const"
	t "bot/src/utils/types"
	"fmt"
	"log"
)

func NotifyAboutPaymentExpiration(bot *bot.Bot) {
	payments, err := db.Query.GetPaymentsEndingSoon(bot.Ctx)

	if err != nil {
		log.Println("Error getting payments ending soon:", err)
		return
	}

	for _, payment := range payments {
		message := `Привет!
Наши занятия подходят к завершению — впереди остаются последние 3 дня.
Надеюсь, ты получила удовольствие от практики и стала чуть лучше чувствовать и понимать своё тело.

Сейчас хорошее время оглянуться назад и спокойно проанализировать проделанную работу.
Обязательно напиши @vialettochka в личные сообщения, чтобы мы могли отследить твой результат и изменения.

Если ты чувствуешь, что не готова останавливаться и хочешь продолжить занятия, есть возможность продления:
<b>+ один дополнительный месяц за 100 ILS</b>`

		bot.SendMessage(t.Message{
			ChatId:    payment.UserID_2,
			Text:      message,
			ParseMode: "html",
			ReplyMarkup: &t.InlineKeyboardMarkup{
				InlineKeyboard: [][]t.InlineKeyboardButton{
					{
						{
							Text:         cnst.ExtendPayment,
							CallbackData: cnst.ExtendPayment,
						},
					},
				},
			},
		})

		adminMessage := fmt.Sprintf(
			"Subscription expiring soon:\nUsername: @%s\nName: %s %s",
			payment.Username,
			payment.FirstName,
			payment.LastName,
		)

		bot.SendMessage(t.Message{
			ChatId: 833382946,
			Text:   adminMessage,
		})

		err = db.Query.MarkPaymentAsNotified(bot.Ctx, payment.ID)
		if err != nil {
			log.Printf("Error marking payment %d as notified: %v\n", payment.ID, err)
		}
	}
}

package commands

import (
	"bot/src/bot"
	"bot/src/pizda/db"
	t "bot/src/utils/types"
)

func Start(bot *bot.Bot, u t.Update) {
	if u.FromChat() == nil || u.Message.Text == "/start" {
		var user t.User


		if u.FromChat() == nil {
			user = u.MyChatMember.From
		} else {
			user = *u.Message.From
		}

		SendKeyboard(bot, user.ID, "Привет. Меня зовут Виолетта. Я йога-терапевт в области женского здоровья. Моя программа направлена конкретно на восстановление и поддержание гинекологической системы женщины, индивидуальную работу с конкретными проблемами организма и коррекцию психо-эмоционального состояния.\n\nВ отличие от классической йоги, которая ориентирована на комплексное развитие тела и сознания мужчин и женщин, здесь я предлагаю более мягкий, адаптированный и целенаправленный подход, учитывающий состояние здоровья и цели молодых женщин. Это не просто тренировки, а выстроенная система, которая приведет тебя к изменениям в движении, дыхании, спорте.")

		db.Query.UpsertUser(bot.Ctx, db.UpsertUserParams{
			TgID:   user.ID,
			FirstName: user.FirstName,
			LastName: user.LastName,
			Username: user.UserName,
		})
	}
}

func pay(bot *bot.Bot, u t.Update) {
	bot.SendMessage(t.Message{
		ChatId: u.FromChat().ID,
		Text:   "Выберите удобный способ оплаты",
		ReplyMarkup: &t.InlineKeyboardMarkup{
			InlineKeyboard: [][]t.InlineKeyboardButton{
				{
					{
						Text:         "Для Израиля 🇮🇱",
						CallbackData: string(db.PizdaPaymentMethodBIT),
					},
					{
						Text:         "Для России 🇷🇺",
						CallbackData: string(db.PizdaPaymentMethodMIR),
					},
				},
			},
		},
	})
}

func SendKeyboard(bot *bot.Bot, chatId int64, text string) {
	var keyboard [][]t.KeyboardButton
	var pair []t.KeyboardButton

	for i := range utils.Keyboard {
		if len(pair) == 2 {
			keyboard = append(keyboard, slices.Clone(pair))
			pair = pair[:0]
		}

		pair = append(pair, t.KeyboardButton{
			Text: utils.Keyboard[i],
		})
	}
	keyboard = append(keyboard, pair)
	replyKeyboard := t.ReplyKeyboardMarkup{
		Keyboard:       keyboard,
		ResizeKeyboard: true,
	}

	msg := t.Message{
		Text:        text,
		ChatId:      chatId,
		ReplyMarkup: &replyKeyboard,
	}

	bot.SendMessage(msg)
}

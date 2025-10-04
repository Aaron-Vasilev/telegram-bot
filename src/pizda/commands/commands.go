package commands

import (
	"bot/src/bot"
	"bot/src/common"
	"bot/src/pizda/db"
	"bot/src/pizda/utils/const"
	t "bot/src/utils/types"
)

func Start(bot *bot.Bot, u t.Update) {
	if u.Message == nil || u.Message.Text == "/start" {
		var user t.User

		if u.FromChat() == nil {
			user = u.MyChatMember.From
		} else {
			user = *u.Message.From
		}

		// 		payments, err := db.Query.IfUserPays(bot.Ctx, user.ID)

		// 		if err != nil {
		// 			bot.Error("p start error: " + err.Error())
		// 		}

		msg := common.GenerateKeyboardMsg(
			user.ID,
			cnst.SaleKeyboard,
			"Привет. Меня зовут Виолетта. Я йога-терапевт в области женского здоровья.Моя программа направлена конкретно на восстановление и поддержание гинекологической системы женщины, индивидуальную работу с конкретными проблемами организма и коррекцию психо-эмоционального состояния.\n\nВ отличие от классической йоги, которая ориентирована на комплексное развитие тела и сознания мужчин и женщин, здесь я предлагаю более мягкий, адаптированный и целенаправленный подход, учитывающий состояние здоровья и цели молодых женщин. Это не просто тренировки, а выстроенная система, которая приведет тебя к изменениям в движении, дыхании, спорте.",
		)

		bot.SendMessage(msg)

		db.Query.UpsertUser(bot.Ctx, db.UpsertUserParams{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Username:  user.UserName,
		})
	}
}

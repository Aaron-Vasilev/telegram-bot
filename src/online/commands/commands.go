package commands

import (
	"bot/src/bot"
	"bot/src/common"
	"bot/src/online/db"
	cnst "bot/src/online/utils/const"
	t "bot/src/utils/types"
)

func Start(bot *bot.Bot, u t.Update) {
	var user t.User

	if u.FromChat() == nil {
		user = u.MyChatMember.From
	} else {
		user = *u.Message.From
	}

	db.Query.UpsertUser(bot.Ctx, db.UpsertUserParams{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Username:  user.UserName,
	})

	bot.SendMessage(
		common.GenerateKeyboardMsg(user.ID, cnst.Keyboard, cnst.GreetingMsg),
	)
}

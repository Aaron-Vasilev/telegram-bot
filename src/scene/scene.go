package scene

import (
	"bot/src/bot"
	"bot/src/controller"
	"bot/src/utils"
	t "bot/src/utils/types"
	"context"
	"database/sql"
	"fmt"
)

type SceneState struct {
	Scene string
	Stage int
}

func GenTokenScene(ctx context.Context, bot *bot.Bot, db *sql.DB, u t.Update) context.Context {
	userId, _ := utils.UserIdFromUpdate(u) 
	state, ok := ctx.Value(userId).(SceneState)

	if !ok {
		bot.Error(fmt.Sprintf("No scene for the user: %d",  userId))
		return ctx
	}

	switch state.Stage {
	case 1:
		buttons := [][]t.InlineKeyboardButton{
			{
				{
					Text: "Once a week",
					CallbackData: "1",
				},
			},
		}

		bot.SendMessage(t.Message {
			Text: "Membership for how many days in a week?",
			ChatId: userId,
			ReplyMarkup: &t.InlineKeyboardMarkup{
				InlineKeyboard: buttons,
			},

		})
	case 2:
		uuidStr := controller.CreateToken(db, u.CallbackQuery.Data)

		bot.SendText(userId, uuidStr)
	}

	state.Stage++
	return context.WithValue(ctx, userId, nil)
}

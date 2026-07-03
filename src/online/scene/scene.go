package scene

import (
	"bot/src/bot"
	"bot/src/common"
	"bot/src/online/db"
	cnst "bot/src/online/utils/const"
	"bot/src/utils"
	t "bot/src/utils/types"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func RegisterScenes(bot *bot.Bot) {
	bot.RegisterScene(cnst.AssignSubscription, assignSubscription)
}

func assignSubscription(bot *bot.Bot, u t.Update) {
	adminId, _ := utils.UserIdFromUpdate(u)
	state, ok := bot.GetCtxValue(adminId)

	if !ok {
		bot.Error(fmt.Sprintf("No scene for the user: %d", adminId))
		bot.EndCtx(adminId)
		return
	}

	switch state.Stage {
	case 1:
		bot.SendMessage(t.Message{
			ChatId:    adminId,
			Text:      "Пришли мне ник, фамилию или имя пользователя",
			ParseMode: "html",
		})
	case 2:
		if u.Message == nil {
			bot.SendText(adminId, utils.WrongMsg)
			bot.EndCtx(adminId)
			return
		}

		if !sendUserList(bot, adminId, u.Message.Text) {
			return
		}
	case 3:
		if u.Message == nil {
			bot.SendText(adminId, utils.WrongMsg)
			bot.EndCtx(adminId)
			return
		}

		payerId, err := strconv.ParseInt(u.Message.Text, 10, 64)
		if err != nil {
			bot.SendText(adminId, "It's not an ID🔫")
			bot.EndCtx(adminId)
			return
		}

		starts := time.Now()
		active, err := db.Query.GetActiveSubscription(bot.Ctx, payerId)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			bot.SendText(adminId, utils.WrongMsg)
			bot.Error("assignSubscription active lookup: " + err.Error())
			bot.EndCtx(adminId)
			return
		}
		if err == nil {
			starts = active.Ends
		}
		ends := starts.AddDate(0, 1, 0)

		_, err = db.Query.CreateSubscription(bot.Ctx, db.CreateSubscriptionParams{
			UserID:   payerId,
			Starts:   starts,
			Ends:     ends,
			IsManual: true,
		})

		if err != nil {
			bot.SendText(adminId, utils.WrongMsg)
			bot.Error("assignSubscription insert: " + err.Error())
			bot.EndCtx(adminId)
			return
		}

		bot.SendText(adminId, "Success 🌋🧯")
		bot.SendMessage(
			common.GenerateKeyboardMsg(
				payerId,
				cnst.Keyboard,
				fmt.Sprintf("Подписка активна до <b>%s</b> 🎊", ends.Format("02-01-06")),
			),
		)

		bot.EndCtx(adminId)
		return
	}

	bot.NextCtx(adminId)
}

func sendUserList(bot *bot.Bot, adminId int64, search string) bool {
	users, err := db.Query.FindUsersByName(bot.Ctx, pgtype.Text{String: search, Valid: true})

	if err != nil {
		bot.Error("find users by name error: " + err.Error())
	}

	if len(users) == 0 {
		bot.SendText(adminId, "There are no users like: "+search)
		bot.EndCtx(adminId)
		return false
	}

	text := ""
	for _, user := range users {
		username := ""
		if user.Username != "" {
			username = "@" + user.Username
		}
		text += fmt.Sprintf("%s %s %s ID = %d\n", user.FirstName, user.LastName, username, user.ID)
	}

	bot.SendText(adminId, text)
	bot.SendText(adminId, "Send back the ID of the user")

	return true
}

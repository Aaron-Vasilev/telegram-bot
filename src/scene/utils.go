package scene

import (
	"bot/src/bot"
	"bot/src/controller"
	"database/sql"
	"fmt"
)

func sendUserList(ctx *Ctx, db *sql.DB, bot *bot.Bot, userID int64, search string) {
	users, err := controller.FindUsersByName(db, search)

	if err != nil {
		bot.Error("find users by name error:" + err.Error())
	}

	if len(users) == 0 {
		bot.SendText(userID, "There are no users like: "+search)
		ctx.End(userID)
		return
	}

	for i := range users {
		userName := ""

		if users[i].Username != "" {
			userName = "@" + users[i].Username
		}
		bot.SendText(userID, fmt.Sprintf("%s %s ID = %d", users[i].Name, userName, users[i].ID))
	}

	bot.SendText(userID, "Send back the ID of the user you want to assign a membership")
}

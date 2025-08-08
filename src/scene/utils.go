package scene

import (
	"bot/src/bot"
	"bot/src/db"
	"fmt"
)

func sendUserList(bot *bot.Bot, userID int64, search string) {
	users, err := db.Query.FindUsersByName(bot.Ctx, search)

	if err != nil {
		bot.Error("find users by name error:" + err.Error())
	}

	if len(users) == 0 {
		bot.SendText(userID, "There are no users like: "+search)
		bot.EndCtx(userID)
		return
	}

	for i := range users {
		userName := ""

		if users[i].Username.Valid {
			userName = "@" + users[i].Username.String
		}
		bot.SendText(userID, fmt.Sprintf("%s %s ID = %d", users[i].Name, userName, users[i].ID))
	}

	bot.SendText(userID, "Send back the ID of the user you want to assign a membership")
}

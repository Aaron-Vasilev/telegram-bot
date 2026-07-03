package action

import (
	"bot/src/bot"
	"bot/src/common"
	"bot/src/online/db"
	"bot/src/utils"
	"fmt"
	"log"
)

func NotifyAboutSubscriptionEnds(bot *bot.Bot) {
	subs, err := db.Query.GetManualSubscriptionsEndingTomorrow(bot.Ctx)
	if err != nil {
		log.Println("Error getting subs ending tomorrow:", err)
		return
	}

	for _, sub := range subs {
		bot.SendHTML(
			sub.UserID,
			"Твоя подписка на онлайн йога-клуб заканчивается <b>завтра</b> 🌸\n\nЧтобы продолжить, напиши @vialettochka",
		)

		username := ""
		if sub.Username != "" {
			username = "@" + sub.Username
		}
		bot.SendHTML(common.AdminChatID(), fmt.Sprintf(
			"Подписка заканчивается завтра:\n%s %s %s\nID: <code>%d</code>",
			sub.FirstName, sub.LastName, username, sub.UserID,
		))

		if err := db.Query.MarkSubscriptionNotified(bot.Ctx, sub.ID); err != nil {
			log.Printf("Error marking sub %d as notified: %v\n", sub.ID, err)
		}
	}
}

func KickExpiredUsers(bot *bot.Bot) {
	ids, err := db.Query.GetExpiredUsers(bot.Ctx)
	if err != nil {
		log.Println("Error getting expired users:", err)
		return
	}

	for _, id := range ids {
		bot.KickChatMember(utils.ONLINE_CHAT_ID, id)
		bot.SendHTML(id, "Твоя подписка закончилась. Ты был удалён из чата. Чтобы вернуться — оформи новую подписку 🔄")
	}

	if len(ids) > 0 {
		bot.SendHTML(common.AdminChatID(), fmt.Sprintf("Удалено %d пользователей без активной подписки", len(ids)))
	}
}

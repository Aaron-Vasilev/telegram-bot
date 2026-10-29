package handler

import (
	"bot/src/bot"
	"bot/src/common"
	"bot/src/online/commands"
	"bot/src/online/db"
	"bot/src/online/payment"
	cnst "bot/src/online/utils/const"
	"bot/src/utils"
	t "bot/src/utils/types"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5"
)

func HandleUpdate(bot *bot.Bot, u t.Update) {
	if u.CallbackQuery == nil && (u.FromChat() == nil || u.Message == nil || strings.HasPrefix(u.Message.Text, "/")) {
		handleMenu(bot, u)
		return
	}

	userId, updateWithCallbackQuery := utils.UserIdFromUpdate(u)
	_, inScene := bot.GetCtxValue(userId)

	if inScene {
		bot.HandleScene(u)
		return
	}

	if updateWithCallbackQuery {
		handleCallbackQuery(bot, u)
		return
	}

	if bot.IfTextScene(u.Message.Text) {
		bot.StartScene(u, u.Message.Text)
		return
	}

	if slices.Contains(cnst.Keyboard, u.Message.Text) {
		handleKeyboard(bot, u)
		return
	}

	if utils.IsAdmin(userId) {
		handleAdminCmd(bot, u)
	}
}

func HandleUpdates(bot *bot.Bot, updates []t.Update) {
	for _, update := range updates {
		HandleUpdate(bot, update)
		bot.Offset = update.UpdateID + 1
	}
}

func handleMenu(bot *bot.Bot, u t.Update) {
	commands.Start(bot, u)

	// t.me/<bot>?start=pay from the payment page on the site
	if u.Message != nil && u.Message.Text == "/start pay" {
		sendPurchaseLink(bot, u.Message.From.ID)
	}
}

func handleKeyboard(bot *bot.Bot, u t.Update) {
	switch u.Message.Text {
	case cnst.Subscription:
		sendSubscriptionInfo(bot, u.FromChat().ID)
	case cnst.Purchase:
		sendPurchaseLink(bot, u.FromChat().ID)
	}
}

func handleCallbackQuery(bot *bot.Bot, u t.Update) {
	switch u.CallbackData() {
	case cnst.Purchase:
		sendPurchaseLink(bot, u.FromChat().ID)
	case cnst.Subscription:
		sendSubscriptionInfo(bot, u.FromChat().ID)
	}
}

func handleAdminCmd(bot *bot.Bot, u t.Update) {
	if u.Message == nil {
		return
	}

	switch u.Message.Text {
	case "ADMIN":
		bot.SendMessage(common.GenerateKeyboardMsg(u.Message.From.ID, cnst.AdminKeyboard, "Admin Keyboard"))
	case "USER":
		bot.SendMessage(common.GenerateKeyboardMsg(u.Message.From.ID, cnst.Keyboard, "User Keyboard"))
	}
}

func sendSubscriptionInfo(bot *bot.Bot, userId int64) {
	sub, err := db.Query.GetActiveSubscription(bot.Ctx, userId)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			bot.SendMessage(t.Message{
				ChatId: userId,
				Text:   cnst.NoActiveSubMsg,
				ReplyMarkup: &t.InlineKeyboardMarkup{
					InlineKeyboard: [][]t.InlineKeyboardButton{
						{{Text: cnst.Purchase, CallbackData: cnst.Purchase}},
					},
				},
			})
		} else {
			bot.Error("get active sub: " + err.Error())
		}
		return
	}

	// Manual subscriptions are granted by the admin and never charged.
	endsLabel := "Следующее списание"
	if sub.IsManual || !sub.IsActive {
		endsLabel = "Действует до"
	}

	bot.SendHTML(
		userId,
		fmt.Sprintf(
			"Подписка 🎟️\nНачалась: <b>%s</b>\n%s: <b>%s</b>",
			sub.Starts.Format("02-01-06"),
			endsLabel,
			sub.Ends.Format("02-01-06"),
		),
	)
}

func sendPurchaseLink(bot *bot.Bot, userId int64) {
	var buttons [][]t.InlineKeyboardButton
	for _, p := range payment.Providers() {
		link, err := p.PaymentLink(userId)
		if err != nil {
			bot.Error(string(p.Name()) + " payment link: " + err.Error())
			continue
		}
		buttons = append(buttons, []t.InlineKeyboardButton{{Text: p.Button(), URL: &link}})
	}

	if len(buttons) == 0 {
		bot.SendText(userId, cnst.ErrorMsg)
		return
	}

	bot.SendMessage(t.Message{
		ChatId:      userId,
		Text:        "Выбери способ оплаты, чтобы оформить подписку 👇",
		ReplyMarkup: &t.InlineKeyboardMarkup{InlineKeyboard: buttons},
	})
}

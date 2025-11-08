package handler

import (
	"bot/src/bot"
	"bot/src/common"
	"bot/src/pizda/commands"
	"bot/src/pizda/db"
	cnst "bot/src/pizda/utils/const"
	"bot/src/utils"
	t "bot/src/utils/types"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
)

func handleCallbackQuery(bot *bot.Bot, u t.Update) {
	text := u.CallbackData()

	if string(db.PizdaPaymentMethodBIT) == text || string(db.PizdaPaymentMethodMIR) == text {
		if string(db.PizdaPaymentMethodBIT) == text {
			bot.SendMessage(t.Message{
				ChatId:    u.FromChat().ID,
				ParseMode: "html",
				Text:      "Сделай перевод через BIT по номеру телефона:\n<b>0534257328</b>\n\nИли банковский перевод:\n Кому: <b>ארון וויולטה וסילב</b>\n Банк: <b>12</b> (hapoalim)\n Сниф: <b>729</b>\n Номер счёта: <b>86676</b>",
			})
		} else {
			bot.SendMessage(t.Message{
				ChatId:    u.FromChat().ID,
				ParseMode: "html",
				Text:      "Отправь мне бансковский перевод на Тинькофф по номеру телефона:\n<b>+79160824901</b>",
			})
		}
		bot.SendMessage(t.Message{
			ChatId: u.FromChat().ID,
			Text:   "A потом перешли @vialettochka скрин с переводом☺️",
		})
	} else if text == cnst.TestTraining {
		sendTestTraining(bot, u.FromChat().ID)
	} else if text == cnst.Programm {
		sendProgramm(bot, u.FromChat().ID)
	} else if text == cnst.Whom {
		sendToWhom(bot, u.FromChat().ID)
	} else if text == cnst.Purchase {
		purchase(bot, u.FromChat().ID)
	} else if text == cnst.Prices {
		sendPrices(bot, u.FromChat().ID)
	} else if text == cnst.Video {
		sendVideo(bot, u)
	}
}

func HandleUpdate(bot *bot.Bot, u t.Update) {
	if u.CallbackQuery == nil && (u.FromChat() == nil || (u.Message == nil || strings.HasPrefix(u.Message.Text, "/"))) {
		handleMenu(bot, u)

		return
	}

	userId, updateWithCallbackQuery := utils.UserIdFromUpdate(u)
	_, ok := bot.GetCtxValue(userId)

	if ok {
		handleScene(bot, u)
	} else if updateWithCallbackQuery {
		handleCallbackQuery(bot, u)
	} else if bot.IfTextScene(u.Message.Text) {
		bot.StartScene(u, u.Message.Text)
	} else if slices.Contains(cnst.SaleKeyboard, u.Message.Text) || slices.Contains(cnst.PayKeyboard, u.Message.Text) {
		handleKeyboard(bot, u)
	} else if utils.IsAdmin(userId) {
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
}

func handleKeyboard(bot *bot.Bot, u t.Update) {
	key := u.Message.Text
	userId := u.FromChat().ID

	switch key {
	case cnst.Whom:
		sendToWhom(bot, userId)
		return
	case cnst.Purchase:
		purchase(bot, u.FromChat().ID)
		return
	case cnst.Programm:
		sendProgramm(bot, u.FromChat().ID)
		return
	case cnst.TestTraining:
		sendTestTraining(bot, u.FromChat().ID)
		return
	case cnst.Prices:
		sendPrices(bot, u.FromChat().ID)
		return
	}

	payment, err := db.Query.GetValidPayment(bot.Ctx, userId)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			bot.SendMessage(common.GenerateKeyboardMsg(u.Message.From.ID, cnst.SaleKeyboard, "Клавиатура пользователя"))
		} else {
			bot.Error("is paying user error: " + err.Error())
		}
		return
	}

	switch key {
	case cnst.Subscription:
		sendSubscription(bot, payment)
	case cnst.Lessons:
		sendLessons(bot, userId)
	}
}

func sendSubscription(bot *bot.Bot, payment db.PizdaPayment) {
	start, end := formatDateRange(payment.Period)

	bot.SendHTML(
		payment.UserID,
		cnst.Subscription+"\nНачалась: <b>"+start+"</b>\nЗакончится: <b>"+end+"</b>",
	)
}

func handleAdminCmd(bot *bot.Bot, u t.Update) {
	if u.Message == nil {
		return
	}
	cmd := u.Message.Text

	if bot.IfTextScene(u.Message.Text) {
		bot.StartScene(u, cmd)
		return
	}

	var msg t.Message

	switch cmd {
	case "ADMIN":
		msg = common.GenerateKeyboardMsg(u.Message.From.ID, cnst.AdminKeyboard, "Admin Keyboard")
	case "USER":
		msg = common.GenerateKeyboardMsg(u.Message.From.ID, cnst.SaleKeyboard, "User Keyboard")
	}

	if u.Message.Video != nil {
		bytes, _ := json.MarshalIndent(u.Message.Video, "", "\t")
		str := string(bytes)
		bot.SendText(362575139, str)
	} else if cmd != "" {
		bot.SendMessage(msg)
	}
}

func handleScene(bot *bot.Bot, u t.Update) {
	bot.HandleScene(u)
}

func sendProgramm(bot *bot.Bot, userId int64) {
	media := []t.InputMediaPhoto{
		{
			BaseInputMedia: t.BaseInputMedia{
				Type:    "photo",
				Media:   "https://bot-telega.s3.il-central-1.amazonaws.com/programm_1.png",
				Caption: "Программа моего курса 📜",
			},
		},
		{
			BaseInputMedia: t.BaseInputMedia{
				Type:  "photo",
				Media: "https://bot-telega.s3.il-central-1.amazonaws.com/programm_2.png",
			},
		},
		{
			BaseInputMedia: t.BaseInputMedia{
				Type:  "photo",
				Media: "https://bot-telega.s3.il-central-1.amazonaws.com/programm_3.png",
			},
		},
	}

	bot.SendMediaGroup(t.Message{
		ChatId: userId,
		Media:  media,
	})
	bot.SendMessage(t.Message{
		ChatId: userId,
		Text:   "Подходим ли мы друг другу👩‍❤️‍💋‍👩?",
		ReplyMarkup: &t.InlineKeyboardMarkup{
			InlineKeyboard: [][]t.InlineKeyboardButton{
				{
					{
						Text:         "Пройти пробную тренировку🤸‍♀️",
						CallbackData: cnst.TestTraining,
					},
				},
			},
		},
	})
}

func sendToWhom(bot *bot.Bot, userId int64) {
	bot.SendMessage(t.Message{
		ChatId:    userId,
		Text:      "Тебе подходит этот курс, если <b>совпадает хоть один</b> из этих пунктов✅:\n\n1. Ты <b>планируешь беременность сейчас или через 1-2 года</b> и хочешь подготовить себя к этому. Чтобы уменьшить риски и перенести этот сложнейший процесс для организма максимально мяго, и насколько это возможно быстро восстановиться - нужно начать заниматься уже сейчас. Если ты хочешь здоровые: беременность, роды, послеродовой период - йога для беременных тут не поможет. Нужна большая комплексная работа с телом, чтобы войти в этот период не просто с сильным выстроенным фундаментом, с внушительным запасом внутреннего ресурса\n\n2. Ты чувствуешь, что тебе нужно восстановление, заняться собой и своим здоровьем. <b>ТЯЖЕЛЫЙ ПМС</b> каждый месяц? У тебя есть <b>нарушения цикла или заболевания по-женски</b>? Возможно, у тебя слабые мышцы тазового дна, опущение или диастаз, если <b>ты уже рожала</b>. Лишний вес или метаболический синдром? Поликистоз или миомы, новообразования? Для составления практик этом курсе я использовала самые современные протоколы, исследования, консультировалась с гинекологами и проктологами. Поэтому каждая тренировка терапевтирует или профилактирует заболевания\n\n3. Ты опытная спортсменка в спорт-зале или любительница, ведущая активный образ жизни. Ты уже разбираешься в йоге и тренировках, <b>обожаешь изучать свое тело</b>, но <b>хочешь еще углубить свои знания</b> в области женского здоровья, чтобы разнообразить свою активность новыми техниками, которые сделают тебя еще более здоровой, красивой. Хочешь улучшить интимную жизнь, добавить мягких и спокойных движений, <b>снизить психо-эмоциональное напряжение, укрепить нервную систему</b>\n\n4. Тебе <b>нужен индивидуальный и системный подход</b> в тренировках, <b>поддерживающий наставник или друг</b>. Я планирую личную работу с каждой участницей: анкетирование, консультация с обсуждением особенностей твоего здоровья, индивидуальный план в зависимости от твоей цели, ответы на вопросы и поддержка. Ты получаешь персонального тренера. Заинтересованного в твоем результате",
		ParseMode: "html",
		ReplyMarkup: &t.InlineKeyboardMarkup{
			InlineKeyboard: [][]t.InlineKeyboardButton{
				{
					{
						Text:         cnst.Programm,
						CallbackData: cnst.Programm,
					},
				},
			},
		},
	})
}

func sendTestTraining(bot *bot.Bot, chatId int64) {
	msg := t.Message{
		ChatId:         chatId,
		ParseMode:      "html",
		ProtectContent: true,
		Video: &t.CustomVideo{
			FileId:   "BAACAgIAAxkBAAIBmmkNAzXDBISkgMPEZrEgzCH0iwsOAAJwjgACQgpoSN7jz9up7CqcNgQ",
			IsString: true,
		},
		Caption: "Посмотри это видео чтобы понять подходим ли мы друг-другу по вайбу😎\nЭто видео средней сложности. В курсе, мы начнём с простых практик и будем двигаться в сторону более продвинутых💪",
		ReplyMarkup: &t.InlineKeyboardMarkup{
			InlineKeyboard: [][]t.InlineKeyboardButton{
				{
					{
						Text:         cnst.Prices,
						CallbackData: cnst.Prices,
					},
				},
			},
		},
	}

	bot.SendVideoById(msg)
}

func purchase(bot *bot.Bot, chatId int64) {
	bot.SendMessage(t.Message{
		ChatId: chatId,
		Text:   "Выберите удобный способ оплаты",
		ReplyMarkup: &t.InlineKeyboardMarkup{
			InlineKeyboard: [][]t.InlineKeyboardButton{
				{
					{
						Text:         "Для Израиля 🇮🇱\n(Bit, банковский перевод)",
						CallbackData: string(db.PizdaPaymentMethodBIT),
					},
				},
				{
					{
						Text:         "Для России 🇷🇺\n(Tinkoff)",
						CallbackData: string(db.PizdaPaymentMethodMIR),
					},
				},
			},
		},
	})
}

func sendPrices(bot *bot.Bot, chatId int64) {
	msg := t.Message{
		ChatId: chatId,
		Media: []t.InputMediaPhoto{
			{
				BaseInputMedia: t.BaseInputMedia{
					Type:      "photo",
					Media:     "https://bot-telega.s3.il-central-1.amazonaws.com/pizda/plan_1.JPG",
					Caption:   "Цены и тарифы:\n1. \nБазовый - <b>300₪</b> или <b>7500₽</b>, включающий консультацию и поддержку в групповом чате2. Индивидуальный - <b>900₪</b> или <b>22500₽</b>, с личными практиками и ведением",
					ParseMode: "html",
				},
			},
			{
				BaseInputMedia: t.BaseInputMedia{
					Type:  "photo",
					Media: "https://bot-telega.s3.il-central-1.amazonaws.com/pizda/plan_2.jpg",
				},
			},
		},
	}
	bot.SendMediaGroup(msg)
	bot.SendMessage(t.Message{
		ChatId: chatId,
		Text:   "Готова начать🌸?",
		ReplyMarkup: &t.InlineKeyboardMarkup{
			InlineKeyboard: [][]t.InlineKeyboardButton{
				{
					{
						Text:         cnst.Purchase,
						CallbackData: cnst.Purchase,
					},
				},
			},
		}})
}

func formatDateRange(r pgtype.Range[pgtype.Date]) (string, string) {
	formatDate := func(d pgtype.Date) string {
		return d.Time.Format("02-01-06")
	}

	return formatDate(r.Lower), formatDate(r.Upper)
}

func sendLessons(bot *bot.Bot, chatId int64) {
	videos, err := db.Query.GetVideos(bot.Ctx)

	if err != nil {
		bot.SendText(chatId, cnst.ErrorMsg)
		bot.Error(fmt.Sprintf("User: %d %s", chatId, err.Error()))
		return
	}

	keys := t.InlineKeyboardMarkup{
		InlineKeyboard: [][]t.InlineKeyboardButton{},
	}

	for _, v := range videos {
		if v.FileID != "" {
			key := []t.InlineKeyboardButton{
				{
					Text:         v.Name,
					CallbackData: cnst.Video,
				},
			}
			keys.InlineKeyboard = append(keys.InlineKeyboard, key)
		}
	}

	bot.SendMessage(t.Message{
		ChatId:      chatId,
		Text:        "Список уроков📀",
		ReplyMarkup: keys,
	})
}

func sendVideo(bot *bot.Bot, u t.Update) {
	userId := u.FromChat().ID
	isUserPays, _ := isPayingUser(bot, userId)

	if !isUserPays {
		bot.SendMessage(common.GenerateKeyboardMsg(userId, cnst.SaleKeyboard, "Клавиатура пользователя"))
	} else {
		markup, err := inlineKeyboardFromReplyMarkup(u.CallbackQuery.Message.ReplyMarkup)
		if err != nil {
			bot.Error("sendVideo reply markup decode error: " + err.Error())
			return
		}

		if markup == nil || len(markup.InlineKeyboard) == 0 || len(markup.InlineKeyboard[0]) == 0 {
			bot.Error("sendVideo reply markup is empty")
			return
		}

		videoName := markup.InlineKeyboard[0][0].Text
		video, err := db.Query.GetVideoByName(bot.Ctx, videoName)

		if err != nil {
			bot.Error("sendVideo error: " + err.Error())
			return
		}

		bot.SendVideoById(t.Message{
			ChatId:  userId,
			Caption: video.Name,
			Video: &t.CustomVideo{
				FileId:   video.FileID,
				IsString: true,
			},
		})
	}
}

func inlineKeyboardFromReplyMarkup(markup interface{}) (*t.InlineKeyboardMarkup, error) {
	if markup == nil {
		return nil, nil
	}

	switch v := markup.(type) {
	case *t.InlineKeyboardMarkup:
		return v, nil
	case t.InlineKeyboardMarkup:
		return &v, nil
	case map[string]interface{}:
		bytes, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}

		var inlineMarkup t.InlineKeyboardMarkup

		if err := json.Unmarshal(bytes, &inlineMarkup); err != nil {
			return nil, err
		}

		return &inlineMarkup, nil
	default:
		return nil, fmt.Errorf("unexpected reply markup type %T", markup)
	}
}

func isPayingUser(bot *bot.Bot, userId int64) (bool, db.PizdaPayment) {
	payment, err := db.Query.GetValidPayment(bot.Ctx, userId)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, payment
		}
		bot.Error("is paying user error: " + err.Error())
		return false, payment
	}

	return true, payment
}

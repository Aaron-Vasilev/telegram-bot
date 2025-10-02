package common

import (
	t "bot/src/utils/types"
	"slices"
)

func GenerateKeyboardMsg(chatId int64, keys []string, text string) t.Message {
	var keyboard [][]t.KeyboardButton
	var pair []t.KeyboardButton

	for i := range keys {
		if len(pair) == 2 {
			keyboard = append(keyboard, slices.Clone(pair))
			pair = pair[:0]
		}

		pair = append(pair, t.KeyboardButton{
			Text: keys[i],
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

	return msg
}

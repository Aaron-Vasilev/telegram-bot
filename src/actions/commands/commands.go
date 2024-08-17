package commands

import (
	"bot/src/utils"
	t "bot/src/utils/types"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type Bot struct {
	Token string
	IsDebug bool
	Offset int
}

func Call[T any](bot *Bot, method string) T {
	var resData t.Response[T]
	res, err := http.Get("https://api.telegram.org/bot" + bot.Token + method)

	if err != nil {
		fmt.Println("Error making the request:", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if err != nil {
		fmt.Println("Error while ioutil.ReadAll:", err)
	}

	err = json.Unmarshal(body, &resData)
	if err != nil {
		fmt.Println("Error json.Unmarshal:", err)
	}
	if !resData.Ok {
		fmt.Println("Response is not OK")
	}

	if bot.IsDebug {
		s, _ := json.MarshalIndent(resData.Result, "", "\t")
		str := string(s)

		if str != "[]" {
			fmt.Println("Messages received: ", str)
		}
	}

	return resData.Result
}

func Send(bot *Bot, method string, msg t.Message) t.Message {
	var resData t.Response[t.Message]
	jsonData, err := json.Marshal(msg)

	if err != nil {
		fmt.Println("Error while json.Marshal:", err)
	}

	res, err := http.Post(
		"https://api.telegram.org/bot" + bot.Token + method,
		"application/json",
		bytes.NewBuffer(jsonData),
		)

	if err != nil {
		fmt.Println("Error making the request:", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if err != nil {
		fmt.Println("Error while ioutil.ReadAll:", err)
	}

	err = json.Unmarshal(body, &resData)
	if err != nil {
		fmt.Println("Error json.Unmarshal:", err)
	} else if !resData.Ok {
		fmt.Println("Response is not OK, ErrorCode:", resData.ErrorCode, resData.Description)
		fmt.Println(resData.Description)
	} else if bot.IsDebug  {
		s, _ := json.MarshalIndent(resData.Result, "", "\t")
		fmt.Println("Messages received: ", string(s))
	}

	return resData.Result
}

func (bot *Bot) GetMe() t.TBot {
	return Call[t.TBot](bot, "/getMe")
}

func (bot *Bot) GetUpdates() []t.Update {
	offset := strconv.Itoa(bot.Offset)

	return Call[[]t.Update](bot, "/getUpdates?timeout=3&offset=" + offset)
}

func (bot *Bot) SendKeyboard(text string) {
	replyKeyboard := t.ReplyKeyboardMarkup{
		Keyboard: [][]t.KeyboardButton{ 
			{
				{
					Text: utils.Keyboard["Timetable 🗓"],
				},
				{
					Text: utils.Keyboard["Leaderboard 🏆"],
				},
			},
			{
				{
					Text: utils.Keyboard["Profile 🧘"],
				},
				{
					Text: utils.Keyboard["Contact 💌"],
				},
			},
		},
		ResizeKeyboard: true,
	}

	msg := t.Message {
		Text: text,
		ChatId: 362575139,
		ReplyMarkup: &replyKeyboard,
	}

	Send(bot, "/sendMessage", msg)
}

func (bot *Bot) SendMessage(chatId int64, text string) {
	msg := t.Message {
		Text: text,
		ChatId: chatId,
	}

	Send(bot, "/sendMessage", msg)
}

func handleKeyboard(bot *Bot, u t.Update) {
	key := u.Message.Text

	switch key {
	case utils.Timetable: 
		bot.SendMessage(u.Message.Chat.ID, "NIHUA")
	}
}

func handleMessage(bot *Bot, update t.Update) {
	if _, exists := utils.Keyboard[update.Message.Text]; exists {
		handleKeyboard(bot, update)
	}
}

func HandleUpdates(bot *Bot, updates []t.Update) {
	for _, update := range updates {
		handleMessage(bot, update)

		bot.Offset = update.UpdateID + 1
	}
}

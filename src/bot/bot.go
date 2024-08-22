package bot

import (
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

func (bot *Bot) GetMe() t.TBot {
	return Call[t.TBot](bot, "/getMe")
}

func (bot *Bot) GetUpdates() []t.Update {
	offset := strconv.Itoa(bot.Offset)

	return Call[[]t.Update](bot, "/getUpdates?timeout=3&offset=" + offset)
}

func (bot *Bot) SendMessage(msg t.Message) {
	Send(bot, "/sendMessage", msg)
}

func (bot *Bot) SendText(chatId int64, text string) {
	msg := t.Message {
		Text: text,
		ChatId: chatId,
	}

	Send(bot, "/sendMessage", msg)
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
		fmt.Println("Messages SEND: ", string(s))
	}

	return resData.Result
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
			fmt.Println("Messages RECEIVED: ", str)
		}
	}

	return resData.Result
}


package bot

import (
	t "bot/src/utils/types"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
)

type Bot struct {
	Token   string
	IsDebug bool
	Offset  int
}

func (bot *Bot) GetMe() t.TBot {
	return Call[t.TBot](bot, "/getMe")
}

func (bot *Bot) GetUpdates() []t.Update {
	offset := strconv.Itoa(bot.Offset)

	return Call[[]t.Update](bot, "/getUpdates?timeout=1&offset="+offset)
}

func (bot *Bot) SendMessage(msg t.Message) {
	Send(bot, "/sendMessage", msg)
}

func (bot *Bot) SendText(chatId int64, text string) {
	msg := t.Message{
		Text:   text,
		ChatId: chatId,
	}

	Send(bot, "/sendMessage", msg)
}

func (bot *Bot) SendHTML(chatId int64, text string) {
	msg := t.Message{
		Text:      text,
		ChatId:    chatId,
		ParseMode: "html",
	}

	Send(bot, "/sendMessage", msg)
}

func (bot *Bot) SendSticker(chatId int64, stickerId string) {
	msg := t.Message{
		ChatId:  chatId,
		Sticker: stickerId,
	}

	Send(bot, "/sendSticker", msg)
}

func (bot *Bot) SendPhotoById(chatId int64, fileId string) {
	msg := t.Message{
		ChatId:  chatId,
		Photo: fileId,
	}

	Send(bot, "/sendPhoto", msg)
}

func (bot *Bot) SendLocation(chatId int64, lat float32, long float32) {
	msg := t.Message{
		ChatId:  chatId,
        Latitude: lat,
        Longitude: long,
	}

	Send(bot, "/sendLocation", msg)
}


func (bot *Bot) Error(text string) {
	if bot.IsDebug {
		fmt.Fprintf(os.Stdout, "\033[0;31m Error \033[0m %s", text)
	} else {
		userId, err := strconv.ParseInt(os.Getenv("ERROR_CHAT_ID"), 10, 64)

		if err == nil {
			fmt.Println("No ERROR_CHAT_ID")
		} else {
			bot.SendText(userId, text)
		}
	}
}

func Send(bot *Bot, method string, msg t.Message) {
	var resData t.Response[t.Message]
	jsonData, err := json.Marshal(msg)

	if err != nil {
		fmt.Println("Error while json.Marshal:", err)
	}

	res, err := http.Post(
		"https://api.telegram.org/bot"+bot.Token+method,
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	defer res.Body.Close()

	if err != nil {
		bot.Error("Error making the request:" + err.Error())
		return
	}

	body, err := io.ReadAll(res.Body)

	if err != nil {
		bot.Error("Error while ioutil.ReadAll:" + err.Error())
		return
	}

	if bot.IsDebug {
		err = json.Unmarshal(body, &resData)
		if err != nil {
			fmt.Println("Error json.Unmarshal:", err)
		} else if !resData.Ok {
			if resData.ErrorCode != 400 && resData.Description != "Bad Request: chat not found" {
				bot.Error(fmt.Sprintf("Response is not OK, code: %d, %s", resData.ErrorCode, resData.Description))
			}
		} else {
			s, _ := json.MarshalIndent(resData.Result, "", "\t")
			fmt.Println("Messages SEND: ", string(s))
		}
    }
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

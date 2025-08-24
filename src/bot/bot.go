package bot

import (
	t "bot/src/utils/types"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Bot struct {
	Token    string
	IsDebug  bool
	LogLevel int
	Offset   int
	Ctx      context.Context
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

func (bot *Bot) Forward(chatId, fromChatId int64, msgId int) ([]byte, error) {
	msg := t.Message{
		ChatId:     chatId,
		MessageID:  msgId,
		FromChatID: fromChatId,
	}

	return Send(bot, "/forwardMessage", msg)
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
		ChatId: chatId,
		Photo:  fileId,
	}

	Send(bot, "/sendPhoto", msg)
}

func (bot *Bot) SendVideoById(chatId int64, fileId string) {
	msg := t.Message{
		ChatId: chatId,
		Video: &t.CustomVideo{
			FileId:   fileId,
			IsString: true,
		},
	}

	Send(bot, "/sendVideo", msg)
}

func (bot *Bot) SendLocation(chatId int64, lat float32, long float32) {
	msg := t.Message{
		ChatId:    chatId,
		Latitude:  lat,
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
			bot.SendText(userId, text)
		} else {
			log.Fatal("No ERROR_CHAT_ID")
		}
	}
}

func Send[T any](bot *Bot, method string, obj T) ([]byte, error) {
	jsonData, err := json.Marshal(obj)

	if err != nil {
		fmt.Println("Error while json.Marshal:", err)
	}

	res, err := http.Post(
		"https://api.telegram.org/bot"+bot.Token+method,
		"application/json",
		bytes.NewBuffer(jsonData),
	)

	if err != nil {
		bot.Error("Error making the request:" + err.Error())
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if err != nil {
		bot.Error("Error while ioutil.ReadAll:" + err.Error())
		return nil, err
	}

	if res.StatusCode == http.StatusBadRequest && isBotBlocked(body) {
		return nil, t.BotIsBlockedError
	}

	if bot.IsDebug {
		var resData t.Response[t.Message]
		err = json.Unmarshal(body, &resData)

		if err != nil {
			bot.Error("Error json.Unmarshal: " + err.Error())
		} else if !resData.Ok {
			if resData.ErrorCode != 400 && resData.Description != "Bad Request: chat not found" {
				bot.Error(fmt.Sprintf("Response is not OK, code: %d, %s", resData.ErrorCode, resData.Description))
			} else {
				bot.Error("Response is not OK: " + resData.Description)
			}
		} else {
			bytes, _ := json.MarshalIndent(resData.Result, "", "\t")
			fmt.Println("Messages SEND: ", string(bytes))
		}
	}

	return body, nil
}

func Call[T any](bot *Bot, method string) T {
	var resData t.Response[T]
	res, err := http.Get("https://api.telegram.org/bot" + bot.Token + method)

	if err != nil {
		bot.Error("Error making the request: " + err.Error())
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if err != nil {
		bot.Error("Error while ioutil.ReadAll: " + err.Error())
	}

	err = json.Unmarshal(body, &resData)
	if err != nil {
		bot.Error("Error json.Unmarshal: " + err.Error())
	}

	if !resData.Ok {
		bot.Error("Response is not OK " + resData.Description)
	}

	if bot.IsDebug {
		bot, _ := json.MarshalIndent(resData.Result, "", "\t")
		str := string(bot)

		if str != "[]" && str != "null" {
			fmt.Println("Messages RECEIVED: ", str)
		}
	}

	return resData.Result
}

func isBotBlocked(body []byte) bool {
	if strings.Contains(string(body), "Bad Request: chat not found") {
		return true
	}
	return false
}

func (bot *Bot) SendMediaGroup(chatId int64, media []t.InputMediaPhoto) ([]byte, error) {
	msg := t.Message{
		ChatId: chatId,
		Media:  media,
	}

	return Send(bot, "/SendMediaGroup", msg)
}

func (bot *Bot) SendPool(poll t.PollMessage) (t.Message, error) {
	var res t.Response[t.Message]
	body, err := Send(bot, "/sendPoll", poll)

	if err != nil {
		return res.Result, err
	}

	err = json.Unmarshal(body, &res)

	if err != nil {
		bot.Error("Send media group error: " + err.Error())
	}

	return res.Result, nil
}

type SceneState struct {
	Scene string
	Stage int
	Data  interface{}
}

type sceneMap map[int64]SceneState

type Ctx struct {
	Ctx context.Context
}

var sceneKey = "SceneKey"

func (bot *Bot) StartCtx(userId int64, scene string) {
	bot.SetCtxValue(userId, SceneState{
		Scene: scene,
		Stage: 1,
	})
}

func (bot *Bot) EndCtx(userId int64) {
	sMap := bot.Ctx.Value(sceneKey).(sceneMap)

	delete(sMap, userId)

	bot.Ctx = context.WithValue(bot.Ctx, sceneKey, sMap)
}

func (bot *Bot) NextCtx(userId int64) {
	sMap := bot.Ctx.Value(sceneKey).(sceneMap)
	state := sMap[userId]

	state.Stage++
	sMap[userId] = state

	bot.Ctx = context.WithValue(bot.Ctx, sceneKey, sMap)
}

func (bot *Bot) GetCtxValue(userId int64) (SceneState, bool) {
	sMap, _ := bot.Ctx.Value(sceneKey).(sceneMap)
	value, exist := sMap[userId]

	return value, exist
}

func (bot *Bot) SetCtxValue(userId int64, state SceneState) {
	sMap, ok := bot.Ctx.Value(sceneKey).(sceneMap)

	if ok {
		sMap[userId] = state
	} else {
		sMap = sceneMap{userId: state}
	}

	bot.Ctx = context.WithValue(bot.Ctx, sceneKey, sMap)
}

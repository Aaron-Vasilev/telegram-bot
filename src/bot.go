package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type callback func(ctx *Ctx, bot *Bot, u Update)
type menuCb func(bot *Bot, u Update)

type command struct {
	regex string
	cb callback
}

type Bot struct {
	Token    string
	IsDebug  bool
	LogLevel int
	Offset   int

	CBQueries []command
	SceneMap map[string]callback
	KeyboardMap map[string]callback
	MenuMap map[string]menuCb
}

func (bot *Bot) GetMe() TBot {
	return call[TBot](bot, "/getMe")
}

func (bot *Bot) GetUpdates() []Update {
	offset := strconv.Itoa(bot.Offset)

	return call[[]Update](bot, "/getUpdates?timeout=1&offset="+offset)
}

func (bot *Bot) SendMessage(msg Message) {
	send(bot, "/sendMessage", msg)
}

func (bot *Bot) SendText(chatId int64, text string) {
	msg := Message{
		Text:   text,
		ChatId: chatId,
	}

	send(bot, "/sendMessage", msg)
}

func (bot *Bot) SendHTML(chatId int64, text string) {
	msg := Message{
		Text:      text,
		ChatId:    chatId,
		ParseMode: "html",
	}

	send(bot, "/sendMessage", msg)
}

func (bot *Bot) Forward(chatId, fromChatId int64, msgId int) {
	msg := Message{
		ChatId:     chatId,
		MessageID:  msgId,
		FromChatID: fromChatId,
	}

	send(bot, "/forwardMessage", msg)
}

func (bot *Bot) SendSticker(chatId int64, stickerId string) {
	msg := Message{
		ChatId:  chatId,
		Sticker: stickerId,
	}

	send(bot, "/sendSticker", msg)
}

func (bot *Bot) SendPhotoById(chatId int64, fileId string) {
	msg := Message{
		ChatId: chatId,
		Photo:  fileId,
	}

	send(bot, "/sendPhoto", msg)
}

func (bot *Bot) SendLocation(chatId int64, lat float32, long float32) {
	msg := Message{
		ChatId:    chatId,
		Latitude:  lat,
		Longitude: long,
	}

	send(bot, "/sendLocation", msg)
}

func (bot *Bot) Error(text string) {
	if bot.IsDebug {
		fmt.Fprintf(os.Stdout, "\033[0;31m Error \033[0m %s", text)
	} else {
		userId, err := strconv.ParseInt(os.Getenv("ERROR_CHAT_ID"), 10, 64)

		if err == nil {
			bot.SendText(userId, text)
		} else {
			fmt.Println("No ERROR_CHAT_ID")
		}
	}
}

func send(bot *Bot, method string, msg Message) {
	var resData Response[Message]
	jsonData, err := json.Marshal(msg)

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
		return
	}
	defer res.Body.Close()

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

func call[T any](bot *Bot, method string) T {
	var resData Response[T]
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

func (bot *Bot) Launch() {
	ctx := newSceneContext()

	fmt.Println("Launch!")

	for {
		updates := bot.GetUpdates()

		handleUpdates(&ctx, bot, updates)
	}
}

func (bot *Bot) NewCallbackQuery(regex string, cb callback) error {
	if _, err := regexp.Compile(regex); err != nil {
		return err
	}

	bot.CBQueries = append(bot.CBQueries, command{
		regex: regex,
		cb: cb,
	})

	return nil
}

func handleUpdates(c *Ctx, bot *Bot, updates []Update) {
	for _, update := range updates {
		handleUpdate(c, bot, update)

		bot.Offset = update.UpdateID + 1
	}
}

func handleCallbackQuery(ctx *Ctx, bot *Bot, u Update) {
	for _, cmd := range bot.CBQueries {
		if regexp.MustCompile(cmd.regex).Match([]byte(u.CallbackData())) {
			cmd.cb(ctx, bot, u)
		}
	}
}

func (bot *Bot) NewScene(sceneName string, cb callback) {
	bot.SceneMap[sceneName] = cb
}

func (bot *Bot) StartScene(ctx *Ctx, u Update, sceneName string) {
	ctx.SetValue(u.FromChat().ID, SceneState{
		Scene: sceneName,
		Stage: 1,
	})

	bot.SceneMap[sceneName](ctx, bot, u)
}

func handleScene(ctx *Ctx, bot *Bot, u Update) {
	state, _ := ctx.GetValue(u.FromChat().ID)

	bot.SceneMap[state.Scene](ctx, bot, u)
}

func (bot *Bot) NewMenuItem(item string, cb menuCb) {
	bot.MenuMap[item] = cb
}

func handleMenu(bot *Bot, u Update) {
	cb, ok := bot.MenuMap[u.Message.Text]

	if ok {
		cb(bot, u)
	}
}

func handleUpdate(ctx *Ctx, bot *Bot, u Update) {
	if u.FromChat() == nil || (u.Message != nil && strings.HasPrefix(u.Message.Text, "/")) {
		handleMenu(bot, u)

		return
	}

	userId, updateWithCallbackQuery := userIdFromUpdate(u)

	_, ok := ctx.GetValue(userId)

	if ok {
		handleScene(ctx, bot, u)
	} else if updateWithCallbackQuery {
		handleCallbackQuery(ctx, bot, u)
	} else if strings.HasPrefix(u.Message.Text, "/") {
		handleMenu(bot, u)
	}
}

func userIdFromUpdate(u Update) (int64, bool) {
	var userId int64
	updateWithCallbackQuery := u.CallbackQuery != nil

	if updateWithCallbackQuery {
		userId = u.CallbackQuery.From.ID
	} else if u.Message != nil {
		userId = u.Message.From.ID
	}

	return userId, updateWithCallbackQuery
}

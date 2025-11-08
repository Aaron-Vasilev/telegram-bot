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
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

type commandCallback = func(bot *Bot, u t.Update)

type Command struct {
	name        string
	description string
	fn          commandCallback
}

type Bot struct {
	Token       string
	IsDebug     bool
	IsProd      bool
	LogLevel    int
	Offset      int
	Ctx         context.Context
	WebhookPort string
	// Commands    []Command
	Scenes map[string]commandCallback
}

func NewBot(token string) *Bot {
	if token == "" {
		log.Fatalf("No token!")
	}

	ctx := context.Background()
	defer ctx.Done()

	isDebug := os.Getenv("LOG_LEVEL") == "DEBUG"
	isProd := os.Getenv("ENV") == "production"
	var webhookPort string

	if isProd {
		webhookPort = os.Getenv("WEBHOOK_PORT")

		if webhookPort == "" {
			log.Fatalf("WEBHOOK_PORT is not set")
		}
	}

	return &Bot{
		Token:       token,
		Offset:      0,
		IsDebug:     isDebug,
		IsProd:      isProd,
		WebhookPort: webhookPort,
		Ctx:         ctx,
		Scenes:      map[string]commandCallback{},
	}
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
		Photo:  &t.MessagePhoto{FileID: fileId},
	}

	Send(bot, "/sendPhoto", msg)
}

func (bot *Bot) SendPhoto(msg t.Message) {
	Send(bot, "/sendPhoto", msg)
}

func (bot *Bot) SendVideoById(msg t.Message) {
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
				bot.Error("Response is not OK description: " + resData.Description)
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
		bot.Error("Get request to Telegram, response is not OK " + resData.Description)
	}

	if bot.IsDebug {
		bytes, _ := json.MarshalIndent(resData.Result, "", "\t")
		str := string(bytes)

		if str != "[]" && str != "null" {
			fmt.Println("Message RECEIVED: ", str)
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

func (bot *Bot) SendMediaGroup(msg t.Message) ([]byte, error) {
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

func (bot *Bot) GetWebhookInfo() (t.WebhookInfo, error) {
	return Call[t.WebhookInfo](bot, "/getWebhookInfo"), nil
}

func (bot *Bot) CheckWebhookStatus() error {
	info, err := bot.GetWebhookInfo()
	if err != nil {
		return fmt.Errorf("failed to get webhook info: %v", err)
	}

	if bot.IsDebug {
		fmt.Printf("Webhook status: URL=%s, Pending=%d, LastError=%s\n", info.URL, info.PendingUpdateCount, info.LastErrorMessage)
	}

	if info.LastErrorMessage != "" {
		return fmt.Errorf("webhook has errors: %s (date: %d)", info.LastErrorMessage, info.LastErrorDate)
	}

	return nil
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

func (bot *Bot) StartLongPulling(handler func(bot *Bot, updates []t.Update)) {
	for {
		updates := bot.GetUpdates()

		handler(bot, updates)
	}
}

func webhookHandler(bot *Bot, handleUpdate func(bot *Bot, update t.Update)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
			return
		}

		var update t.Update
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			bot.Error(fmt.Sprintf("Failed to decode webhook update: %v", err))

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
			return
		}

		go func() {
			defer func() {
				if r := recover(); r != nil {
					bot.Error(fmt.Sprintf("Panic in webhook handler: %v", r))
				}
			}()

			handleUpdate(bot, update)
		}()

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}
}

func (bot *Bot) StartWebhook(handler func(bot *Bot, update t.Update)) {
	log.Printf("Starting the webhook")

	if err := bot.CheckWebhookStatus(); err != nil {
		log.Printf("Webhook status warning: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", webhookHandler(bot, handler))

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr:    ":" + bot.WebhookPort,
		Handler: mux,
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		fmt.Println("Shutting down webhook server...")

		server.Shutdown(context.Background())
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Webhook server failed: %v", err)
	}
}

// func (bot *Bot) RegisterCommand(commandName, description string, fn commandCallback) {
// 	if !strings.HasPrefix("/", commandName) {
// 		commandName = "/" + commandName
// 	}

// 	bot.Commands = append(bot.Commands, Command{
// 		name: commandName,
// 		description: description,
// 		fn: fn,
// 	})
// }

// func (bot *Bot) HandleCommand(u t.Update) {
// 	i := slices.IndexFunc(bot.Commands, func(c Command) bool { return c.name == u.Message.Text })

// 	if i == -1 {
// 		bot.SendText(u.FromChat().ID, "This command s unregistered")
// 	} else {
// 		bot.Commands[i].fn(bot, u)
// 	}
// }

// func (bot *Bot) SetMyCommands() {
// 	Send(bot, "/setMyCommands", bot.Commands)
// }

func (bot *Bot) RegisterScene(sceneName string, fn commandCallback) {
	bot.Scenes[sceneName] = fn
}

func (bot *Bot) HandleScene(u t.Update) {
	sceneState, exist := bot.GetCtxValue(u.FromChat().ID)
	sceneCb, exist := bot.Scenes[sceneState.Scene]

	if exist {
		sceneCb(bot, u)
	} else {
		bot.EndCtx(u.FromChat().ID)
		bot.Error("Scene doesn't exist: " + sceneState.Scene)
	}
}

func (bot *Bot) StartScene(u t.Update, sceneName string) {
	bot.SetCtxValue(u.FromChat().ID, SceneState{
		Scene: sceneName,
		Stage: 1,
	})

	bot.Scenes[sceneName](bot, u)
}

func (bot *Bot) IfTextScene(sceneName string) bool {
	_, exist := bot.Scenes[sceneName]
	return exist
}

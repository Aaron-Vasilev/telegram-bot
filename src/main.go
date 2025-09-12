package main

import (
	"bot/src/bot"
	"bot/src/cron"
	"bot/src/db"
	"bot/src/handler"
	"bot/src/utils"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"
)

func main() {
	utils.LoadEnv()

	bot := bot.NewBot(os.Getenv("TOKEN"))
	connection := db.ConnectDB(bot)
	defer connection.Close(bot.Ctx)
	cron.Cron(bot)

	if bot.IsProd {
		startWebhookMode(bot)
	} else {
		startPollingMode(bot)
	}
}

func startPollingMode(bot *bot.Bot) {
	for {
		updates := bot.GetUpdates()

		handler.HandleUpdates(bot, updates)
	}
}

func startWebhookMode(bot *bot.Bot) {
	if err := bot.CheckWebhookStatus(); err != nil {
		log.Printf("Webhook status warning: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", handler.WebhookHandler(bot))

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

package main

import (
	"bot/src/bot"
	"bot/src/cron"
	"bot/src/handler"
	"bot/src/scene"
	"bot/src/utils"
	t "bot/src/utils/types"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"database/sql"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/lib/pq"
)

func NewBot(token string) *bot.Bot {
	return &bot.Bot{
		Token:   token,
		Offset:  0,
		IsDebug: os.Getenv("ENV") == "DEBUG",
	}
}

func main() {
	utils.LoadEnv()

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))

	if err = db.Ping(); err != nil {
		log.Fatal("Error connecting to db", err)
	}
	defer db.Close()

	bot := NewBot(os.Getenv("TOKEN"))

	ctx := scene.NewSceneContext()
	cron.Cron(bot, db)

	fmt.Println("Launch!")

	if bot.IsDebug {
		for {
			updates := bot.GetUpdates()

			handler.HandleUpdates(&ctx, bot, db, updates)
		}
	} else {
		lambda.Start(func(c context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
			var update t.Update

			err := json.Unmarshal([]byte(req.Body), &update)

			if err != nil {
				fmt.Println("Failed to parse request body:", err)
				bot.Error(err.Error())

				return events.APIGatewayProxyResponse{StatusCode: 400}, nil
			}

			handler.HandleUpdates(&ctx, bot, db, []t.Update{update})

			return events.APIGatewayProxyResponse{StatusCode: 200}, nil
		})
	}
}

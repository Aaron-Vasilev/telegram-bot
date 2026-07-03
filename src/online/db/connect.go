package db

import (
	"bot/src/bot"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	_ "github.com/lib/pq"
)

var Query *Queries

func ConnectDB(bot *bot.Bot) *pgx.Conn {
	connection, err := pgx.Connect(bot.Ctx, os.Getenv("DATABASE_URL"))
	Query = New(connection)

	if bot.IsDebug {
		fmt.Println("DB is connected")
	}
	if err != nil {
		log.Fatal("Error connecting to db", err)
	}

	return connection
}

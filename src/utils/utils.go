package utils

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() {
  env := os.Getenv("ENV")

  if env == "" { 
    err := godotenv.Load()

    if err != nil {
      log.Fatal("No .env")
    }
  }
}


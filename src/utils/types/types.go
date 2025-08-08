package t

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	BotIsBlockedError = errors.New("Bot is blocked")
)

type Membership struct {
	UserID           int64
	Starts           time.Time
	Ends             time.Time
	Type             int
	LessonsAvailable int
}

type Token struct {
	ID      uuid.UUID
	Type    int
	Created time.Time
	Valid   bool
}

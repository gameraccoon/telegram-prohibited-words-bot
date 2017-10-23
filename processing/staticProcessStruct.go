package processing

import (
	"github.com/gameraccoon/telegram-prohibited-words-bot/chat"
	"github.com/gameraccoon/telegram-prohibited-words-bot/database"
	"github.com/nicksnyder/go-i18n/i18n"
)

type UserState int

const (
	Normal UserState = iota
	WaitingText
	WaitingVariants
	WaitingRules
)

type StaticConfiguration struct {
	Language    string
	Moderators  []int64
	ExtendedLog bool
}

type StaticProccessStructs struct {
	Chat       chat.Chat
	Db         *database.Database
	Trans      i18n.TranslateFunc
	CachedWords map[int64][]string
}

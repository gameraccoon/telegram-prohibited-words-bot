package telegramChat

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type TelegramChat struct {
	bot *tgbotapi.BotAPI
}

func MakeTelegramChat(apiToken string) (bot *TelegramChat, outErr error) {
	newBot, err := tgbotapi.NewBotAPI(apiToken)
	if err != nil {
		outErr = err
		return
	}

	bot = &TelegramChat{
		bot: newBot,
	}

	return
}

func (telegramChat *TelegramChat) GetBot() *tgbotapi.BotAPI {
	return telegramChat.bot
}

func (telegramChat *TelegramChat) GetBotUsername() string {
	return telegramChat.bot.Self.UserName
}

func (telegramChat *TelegramChat) SetDebugModeEnabled(isEnabled bool) {
	telegramChat.bot.Debug = isEnabled
}

func (telegramChat *TelegramChat) SendMessage(chatId int64, message string) {
	msg := tgbotapi.NewMessage(chatId, message)
	msg.ParseMode = "HTML"
	telegramChat.bot.Send(msg)
}

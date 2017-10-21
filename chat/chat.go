package chat

type Chat interface {
	SendMessage(chatId int64, message string)
}

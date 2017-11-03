package chat

type Chat interface {
	SendMessage(chatId int64, message string)
	IsUserAdmin(chatId int64, userId int64) bool
}

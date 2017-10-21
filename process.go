package main

import (
	"bytes"
	"fmt"
	"github.com/gameraccoon/telegram-prohibited-words-bot/processing"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"strings"
)

type ProcessorFunc func(*processing.ProcessData)

type ProcessorFuncMap map[string]ProcessorFunc

type Processors struct {
	Main ProcessorFuncMap
}

func addWordCommand(data *processing.ProcessData) {
	words := strings.Split(data.Message, ",")

	for _, word := range words {
		data.Static.Db.AddProhibitedWord(strings.Trim(word, " \t\n"))
	}

	data.Static.Chat.SendMessage(data.ChatId, data.Static.Trans("success_message"))
}

func removeWordCommand(data *processing.ProcessData) {
	words := strings.Split(data.Message, ",")

	for _, word := range words {
		data.Static.Db.RemoveProhibitedWord(strings.Trim(word, " \t\n"))
	}

	data.Static.Chat.SendMessage(data.ChatId, data.Static.Trans("success_message"))
}

func listOfWordsCommand(data *processing.ProcessData) {
	var buffer bytes.Buffer

	buffer.WriteString(data.Static.Trans("words_list_header"))

	words := data.Static.Db.GetProhibitedWords()

	for _, word := range words {
		buffer.WriteString(fmt.Sprintf(" %s", word))
	}

	data.Static.Chat.SendMessage(data.ChatId, buffer.String())
}

func playerScoresCommand(data *processing.ProcessData) {
	var buffer bytes.Buffer

	buffer.WriteString(data.Static.Trans("users_list_header"))

	_, names, scores := data.Static.Db.GetUsersList()

	for idx, name := range names {
		score := scores[idx]
		buffer.WriteString(fmt.Sprintf("\n%s - %d", name, score))
	}

	data.Static.Chat.SendMessage(data.ChatId, buffer.String())
}

func makeUserCommandProcessors() ProcessorFuncMap {
	return map[string]ProcessorFunc{
		"add_word":    addWordCommand,
		"remove_word": removeWordCommand,
		"words":       listOfWordsCommand,
		"scores":      playerScoresCommand,
	}
}

func processCommandByProcessors(data *processing.ProcessData, processorsMap ProcessorFuncMap) bool {
	processor, ok := processorsMap[data.Command]
	if ok {
		processor(data)
	}

	return ok
}

func processCommand(data *processing.ProcessData, processors *Processors) {
	processed := processCommandByProcessors(data, processors.Main)
	if processed {
		return
	}

	// if we here it means that no command was processed
	data.Static.Chat.SendMessage(data.ChatId, data.Static.Trans("warn_unknown_command"))
}

func getUserName(update *tgbotapi.Update) string {
	user := update.Message.From
	if user != nil {
		if len(user.UserName) > 0 {
			return "@" + user.UserName
		} else {
			return user.FirstName
		}
	} else {
		return "unknown"
	}
}

func calcProhibitedWordsCount(text string, words []string) (count int) {
	for _, word := range words {
		count += strings.Count(text, strings.ToUpper(word))
	}

	return
}

func processPlainMessage(data *processing.ProcessData) {
	// ToDo: cache uppercase words
	words := data.Static.Db.GetProhibitedWords()

	upperText := strings.ToUpper(data.Message)

	fines := calcProhibitedWordsCount(upperText, words)

	if fines > 0 {
		userId := data.Static.Db.GetUserId(data.ChatId, data.UserName)

		data.Static.Db.AddUserScore(userId, fines)

		data.Static.Chat.SendMessage(data.ChatId, fmt.Sprintf("%s: %d\n%s: %d",
			data.Static.Trans("fine_message"),
			fines,
			data.Static.Trans("total_score_message"),
			data.Static.Db.GetUserScore(userId)))
	}
}

func processUpdate(update *tgbotapi.Update, staticData *processing.StaticProccessStructs, processors *Processors) {
	data := processing.ProcessData{
		Static: staticData,
		ChatId: update.Message.Chat.ID,
	}

	message := update.Message.Text

	if strings.HasPrefix(message, "/") {
		commandLen := strings.Index(message, " ")
		if commandLen != -1 {
			data.Command = message[1:commandLen]
			data.Message = message[commandLen+1:]
		} else {
			data.Command = message[1:]
		}

		processCommand(&data, processors)
	} else {
		data.Message = message
		data.UserName = getUserName(update)
		processPlainMessage(&data)
	}
}

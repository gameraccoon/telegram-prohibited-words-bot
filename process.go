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

	_, names, scores := data.Static.Db.GetUsersList()

	for idx, name := range names {
		score := scores[idx]
		buffer.WriteString(fmt.Sprintf("\n%s - %d", name, score))
	}

	data.Static.Chat.SendMessage(data.ChatId, buffer.String())
}

func playerScoresCommand(data *processing.ProcessData) {
}

func makeUserCommandProcessors() ProcessorFuncMap {
	return map[string]ProcessorFunc{
		"add_word":      addWordCommand,
		"remove_word":   removeWordCommand,
		"list_of_words": listOfWordsCommand,
		"scores":        playerScoresCommand,
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

func processPlainMessage(data *processing.ProcessData) {

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
		processPlainMessage(&data)
	}
}

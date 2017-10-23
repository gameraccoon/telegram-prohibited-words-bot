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
		trimmedWord := strings.Trim(word, " \t\n")
		if len(trimmedWord) > 1 {
			data.Static.Db.AddProhibitedWord(data.ChatId, trimmedWord)
		}
	}

	delete(data.Static.CachedWords, data.ChatId)

	data.Static.Chat.SendMessage(data.ChatId, data.Static.Trans("success_message"))
}

func removeWordCommand(data *processing.ProcessData) {
	words := strings.Split(data.Message, ",")

	for _, word := range words {
		data.Static.Db.RemoveProhibitedWord(data.ChatId, strings.Trim(word, " \t\n"))
	}

	delete(data.Static.CachedWords, data.ChatId)

	data.Static.Chat.SendMessage(data.ChatId, data.Static.Trans("success_message"))
}

func listOfWordsCommand(data *processing.ProcessData) {
	var buffer bytes.Buffer

	buffer.WriteString(data.Static.Trans("words_list_header") + "\n")

	words := data.Static.Db.GetProhibitedWords(data.ChatId)

	for _, word := range words {
		buffer.WriteString(fmt.Sprintf("'%s' ", word))
	}

	data.Static.Chat.SendMessage(data.ChatId, buffer.String())
}

func playerScoresCommand(data *processing.ProcessData) {
	var buffer bytes.Buffer

	buffer.WriteString(data.Static.Trans("users_list_header"))

	_, names, scores := data.Static.Db.GetUsersList(data.ChatId)

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
		"score":       playerScoresCommand,
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
			return user.UserName
		} else {
			return user.FirstName
		}
	} else {
		return "unknown"
	}
}

func calcWordsCount(text string, words []string) (count int) {
	removePunctuation := func(r rune) rune {
		if strings.ContainsRune(".,:;\"'!@#$%^&*()_+=/\\<>[]{}~", r) {
			return -1
		} else {
			return r
		}
	}

	processingText := text
	processingText = strings.Map(removePunctuation, processingText)
	textWords := strings.Fields(processingText)

	for _, word := range words {
		upperWord := word
		for _, textWord := range textWords {
			if upperWord == textWord {
				count++
			}
		}
	}

	return
}

func getProhibitedWords(staticData *processing.StaticProccessStructs, chatId int64) []string {
	if _, ok := staticData.CachedWords[chatId]; !ok {
		words := staticData.Db.GetProhibitedWords(chatId)

		upperWords := words
		for _, word := range words {
			upperWords = append(upperWords, strings.ToUpper(word))
		}

		staticData.CachedWords[chatId] = upperWords
	}

	return staticData.CachedWords[chatId]
}

func processPlainMessage(data *processing.ProcessData) {
	// ToDo: cache uppercase words
	words := getProhibitedWords(data.Static, data.ChatId)

	upperText := strings.ToUpper(data.Message)

	fines := calcWordsCount(upperText, words)

	if fines > 0 {
		data.Static.Db.UpdateUser(data.ChatId, data.UserId, data.UserName)

		data.Static.Db.AddUserScore(data.ChatId, data.UserId, fines)

		data.Static.Chat.SendMessage(data.ChatId, fmt.Sprintf("%s: %d\n%s: %d",
			data.Static.Trans("fine_message"),
			fines,
			data.Static.Trans("total_score_message"),
			data.Static.Db.GetUserScore(data.ChatId, data.UserId)))
	}
}

func processUpdate(update *tgbotapi.Update, staticData *processing.StaticProccessStructs, processors *Processors) {
	data := processing.ProcessData{
		Static: staticData,
		ChatId: update.Message.Chat.ID,
		UserId: int64(update.Message.From.ID),
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
		if update.Message.ForwardFrom == nil {
			data.Message = message
			data.UserName = getUserName(update)
			processPlainMessage(&data)
		}
	}
}

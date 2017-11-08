package main

import (
	"bytes"
	"fmt"
	"github.com/gameraccoon/telegram-prohibited-words-bot/processing"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"strconv"
	"strings"
)

type ProcessorFunc func(*processing.ProcessData)

type ProcessorFuncMap map[string]ProcessorFunc

type Processors struct {
	Main ProcessorFuncMap
}

func isSenderAnAdmin(data *processing.ProcessData) bool {
	if data.AllMembersAreAdmins {
		return true
	} else if data.Static.Chat.IsUserAdmin(data.ChatId, data.UserId) {
		return true
	} else {
		return false
	}
}

func addWordCommand(data *processing.ProcessData) {
	if !isSenderAnAdmin(data) {
		data.Static.Chat.SendMessage(data.ChatId, data.Static.Trans("no_authority"))
		return
	}

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
	if !isSenderAnAdmin(data) {
		data.Static.Chat.SendMessage(data.ChatId, data.Static.Trans("no_authority"))
		return
	}

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

func amnestyLastWords(data *processing.ProcessData) {
	if !isSenderAnAdmin(data) {
		data.Static.Chat.SendMessage(data.ChatId, data.Static.Trans("no_authority"))
		return
	}

	count, err := strconv.Atoi(data.Message)
	if err != nil || count < 1 {
		data.Static.Chat.SendMessage(data.ChatId, data.Static.Trans("wrong_count"))
		return
	}

	words, userId := data.Static.Db.RevokeLastUsedWords(data.ChatId, count, data.UserId)

	if len(words) <= 0 && userId == -1 {
		data.Static.Chat.SendMessage(data.ChatId, data.Static.Trans("no_words_amnestied"))
		return
	}

	data.Static.Chat.SendMessage(data.ChatId,
		fmt.Sprintf(data.Static.Trans("amnestied_words_header"),
			data.Static.Db.GetUserName(data.ChatId, userId),
			strings.Join(words, ", "),
		),
	)
}

func makeUserCommandProcessors() ProcessorFuncMap {
	return map[string]ProcessorFunc{
		"add_word":    addWordCommand,
		"remove_word": removeWordCommand,
		"words":       listOfWordsCommand,
		"score":       playerScoresCommand,
		"amnesty":     amnestyLastWords,
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

func findWords(text string, words []string) (foundWords []string) {
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

	for _, knownWord := range words {
		for _, textWord := range textWords {
			if strings.EqualFold(knownWord, textWord) {
				foundWords = append(foundWords, knownWord)
			}
		}
	}

	return
}

func getProhibitedWords(staticData *processing.StaticProccessStructs, chatId int64) []string {
	if cachedWords, ok := staticData.CachedWords[chatId]; ok {
		return cachedWords
	} else {
		cachedWords := staticData.Db.GetProhibitedWords(chatId)
		staticData.CachedWords[chatId] = cachedWords
		return cachedWords
	}
}

func processPlainMessage(data *processing.ProcessData) {
	// ToDo: cache uppercase words
	words := getProhibitedWords(data.Static, data.ChatId)

	usedProhibitedWords := findWords(data.Message, words)

	if len(usedProhibitedWords) > 0 {
		data.Static.Db.UpdateUser(data.ChatId, data.UserId, data.UserName)

		data.Static.Db.AddWordsUsage(data.ChatId, data.UserId, usedProhibitedWords)

		data.Static.Chat.SendMessage(data.ChatId, fmt.Sprintf("%s: %d (%s)\n%s: %d",
			data.Static.Trans("fine_message"),
			len(usedProhibitedWords),
			strings.Join(usedProhibitedWords, ", "),
			data.Static.Trans("total_score_message"),
			data.Static.Db.GetUserScore(data.ChatId, data.UserId),
		))
	}
}

func processUpdate(update *tgbotapi.Update, staticData *processing.StaticProccessStructs, processors *Processors) {
	data := processing.ProcessData{
		Static:              staticData,
		ChatId:              update.Message.Chat.ID,
		UserId:              int64(update.Message.From.ID),
		AllMembersAreAdmins: update.Message.Chat.AllMembersAreAdmins || update.Message.Chat.IsPrivate(),
	}

	message := update.Message.Text

	if strings.HasPrefix(message, "/") {
		commandLen := strings.Index(message, " ")
		if commandLen != -1 {
			data.Command = strings.Split(message[1:commandLen], "@")[0]
			data.Message = message[commandLen+1:]
		} else {
			data.Command = strings.Split(message[1:], "@")[0]
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

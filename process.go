package main

import (
	//"github.com/gameraccoon/telegram-poll-bot/database"
	"github.com/gameraccoon/telegram-poll-bot/processing"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"strings"
)

type ProcessorFunc func(*processing.ProcessData)

type ProcessorFuncMap map[string]ProcessorFunc

type Processors struct {
	Main ProcessorFuncMap
}

/*
func sendResults(staticData *processing.StaticProccessStructs, questionId int64, chatIds []int64) {
	variants := staticData.Db.GetQuestionVariants(questionId)
	answers := staticData.Db.GetQuestionAnswers(questionId)
	answersCount := staticData.Db.GetQuestionAnswersCount(questionId)

	var buffer bytes.Buffer
	buffer.WriteString(staticData.Trans("results_header"))
	buffer.WriteString(fmt.Sprintf("<i>%s</i>", staticData.Db.GetQuestionText(questionId)))

	for i, variant := range variants {
		buffer.WriteString(fmt.Sprintf("\n%s - %d (%d%%)", variant, answers[i], int64(100.0*float32(answers[i])/float32(answersCount))))
	}
	resultText := buffer.String()

	for _, chatId := range chatIds {
		staticData.Chat.SendMessage(chatId, resultText)
	}
}

func addQuestionCommand(data *processing.ProcessData) {
	if data.Static.Db.IsUserBanned(data.UserId) {
		data.Static.Chat.SendMessage(data.ChatId, data.Static.Trans("warn_youre_banned"))
		return
	}
	if !data.Static.Db.IsUserEditingQuestion(data.UserId) {
		data.Static.Db.StartCreatingQuestion(data.UserId)
		data.Static.Db.UnmarkUserReady(data.UserId)
		data.Static.UserStates[data.ChatId] = processing.WaitingText
		data.Static.Chat.SendMessage(data.ChatId, data.Static.Trans("ask_question_text"))
	} else {
		sendEditingGuide(data, dialogManager)
	}
}

func startCommand(data *processing.ProcessData) {
	data.Static.Chat.SendMessage(data.ChatId, data.Static.Trans("hello_message"))
	if !data.Static.Db.IsUserHasPendingQuestions(data.UserId) {
		data.Static.Db.InitNewUserQuestions(data.UserId)
		data.Static.Db.UnmarkUserReady(data.UserId)
		processing.ProcessNextQuestion(data)
	}
}

func lastResultsCommand(data *processing.ProcessData) {
	questions := data.Static.Db.GetLastFinishedQuestions(10)
	for _, questionId := range questions {
		sendResults(data.Static, questionId, []int64{data.ChatId})
	}
}

func myQuestionsCommand(data *processing.ProcessData) {
	questionsIds := data.Static.Db.GetUserLastQuestions(data.UserId, 10)
	finishedQuestionsIds := data.Static.Db.GetUserLastFinishedQuestions(data.UserId, 10)

	finishedQuestionsMap := make(map[int64]bool)
	for _, questionId := range finishedQuestionsIds {
		finishedQuestionsMap[questionId] = true
	}

	for _, questionId := range questionsIds {
		if _, ok := finishedQuestionsMap[questionId]; ok {
			sendResults(data.Static, questionId, []int64{data.ChatId})
		} else {
			data.Static.Chat.SendMessage(data.ChatId, fmt.Sprintf("<i>%s</i>\n%s",
				data.Static.Db.GetQuestionText(questionId),
				getDificientDataForQuestionText(data.Static, questionId),
			))
		}
	}
}*/

func makeUserCommandProcessors() ProcessorFuncMap {
	return map[string]ProcessorFunc{
	//"add_word": addWordCommand,
	//"remove_word": removeWordCommand,
	//"list_of_words": listOfWordsCommand,
	//"scores": playerScoresCommand,
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
		UserId: staticData.Db.GetUserId(update.Message.Chat.ID),
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

package tbotopenai

import (
	"context"
	"errors"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

const (
	minJobID = 100000
	maxJobID = 999999
)

const (
	labelChatGPT    = "ChatGPT"
	labelOpenAI     = "OpenAI"
	labelDreamBooth = "DreamBooth"
)

func (t *TBotOpenAI) processTask(text string, chatID int64) ([]byte, string) {
	command, err := t.clientStates.ClientCommand(chatID)
	if err != nil {
		t.log.Error("Get client command err:", zap.Error(err))
		return nil, ""
	}
	username, err := t.clientStates.ClientUsername(chatID)
	if err != nil {
		t.log.Error("Get client username err:", zap.Error(err))
		return nil, ""
	}
	val, ok := t.taskByCmd.Load(command)
	if !ok {
		return []byte(respBodyUndefinedJob), ""
	}
	f, ok := val.(func(text string, chatID int64) (body []byte, fileName string))
	if !ok {
		return []byte(respBodyUndefinedJob), ""
	}
	body, filename := f(text, chatID)
	var response string
	if filename == "" {
		response = string(body)
	}
	t.writeStats(command, username, text, response)
	return body, filename
}

func (t *TBotOpenAI) processCancelJob(text string, chatID int64) ([]byte, string) {
	jobID, err := strconv.Atoi(text)
	if err != nil {
		t.log.Error("Get jobID err:", zap.Error(err))
		return []byte(respErrBodyInvalidFormatJobID), ""
	}
	if err = t.clientStates.ClientCancelChatGPTJob(jobID, chatID); err == nil {
		return respBodySuccessCancelJob(labelChatGPT, jobID), ""
	}
	if err = t.clientStates.ClientCancelOpenAIJob(jobID, chatID); err == nil {
		return respBodySuccessCancelJob(labelOpenAI, jobID), ""
	}
	if err = t.clientStates.ClientCancelDreamBoothJob(jobID, chatID); err == nil {
		return respBodySuccessCancelJob(labelDreamBooth, jobID), ""
	}
	return respErrBodyJobIsNotExist(jobID), ""
}

func (t *TBotOpenAI) processChatGPT(text string, chatID int64) ([]byte, string) {
	ctx, cancel := context.WithTimeout(context.Background(), t.cfg.ChatGPT.Timeout)
	jobID := randIntByRange(minJobID, maxJobID)
	if err := t.clientStates.ClientAddChatGPTJob(cancel, jobID, chatID); err != nil {
		t.log.Error("Add ChatGPT job err:", zap.Error(err))
		return []byte(respBodySessionIsNotExist), ""
	}
	body, err := t.chatGPTBot.GenerateText(ctx, text)
	if errors.Is(err, context.Canceled) {
		return []byte(respErrBodyJobCanceled), ""
	}
	defer func() {
		if err = t.clientStates.ClientCancelChatGPTJob(jobID, chatID); err != nil {
			t.log.Error("Cancel ChatGPT job err:", zap.Error(err))
		}
	}()
	if err != nil {
		t.log.Error("ChatGPT response err:", zap.Error(err))
		body = []byte(respErrBodyChatGPT)
	}
	return body, ""
}

func (t *TBotOpenAI) processOpenAIText(text string, chatID int64) ([]byte, string) {
	ctx, cancel := context.WithTimeout(context.Background(), t.cfg.OpenAI.Timeout)
	jobID := randIntByRange(minJobID, maxJobID)
	if err := t.clientStates.ClientAddOpenAIJob(cancel, jobID, chatID); err != nil {
		t.log.Error("Add OpenAI job err:", zap.Error(err))
		return []byte(respBodySessionIsNotExist), ""
	}
	body, err := t.openAI.GenerateText(ctx, text)
	if errors.Is(err, context.Canceled) {
		return []byte(respErrBodyJobCanceled), ""
	}
	defer func() {
		if err = t.clientStates.ClientCancelOpenAIJob(jobID, chatID); err != nil {
			t.log.Error("Cancel OpenAI job err:", zap.Error(err))
		}
	}()
	if err != nil {
		t.log.Error("OpenAI response err:", zap.Error(err))
		body = []byte(respErrBodyOpenAI)
	}
	return body, ""
}

func (t *TBotOpenAI) processOpenAIImage(text string, chatID int64) ([]byte, string) {
	ctx, cancel := context.WithTimeout(context.Background(), t.cfg.OpenAI.Timeout)
	jobID := randIntByRange(minJobID, maxJobID)
	if err := t.clientStates.ClientAddOpenAIJob(cancel, jobID, chatID); err != nil {
		t.log.Error("Add OpenAI job err:", zap.Error(err))
		return []byte(respBodySessionIsNotExist), ""
	}
	body, fileName, err := t.openAI.GenerateImage(ctx, text)
	if errors.Is(err, context.Canceled) {
		return []byte(respErrBodyJobCanceled), ""
	}
	defer func() {
		if err = t.clientStates.ClientCancelOpenAIJob(jobID, chatID); err != nil {
			t.log.Error("Cancel OpenAI job err:", zap.Error(err))
		}
	}()
	if err != nil {
		t.log.Error("OpenAI response err:", zap.Error(err))
		body = []byte(respErrBodyOpenAI)
	}
	return body, fileName
}

func (t *TBotOpenAI) processDreamBooth(text string, chatID int64) ([]byte, string) {
	ctx, cancel := context.WithTimeout(context.Background(), t.cfg.DreamBooth.Timeout)
	jobID := randIntByRange(minJobID, maxJobID)
	if err := t.clientStates.ClientAddDreamBoothJob(cancel, jobID, chatID); err != nil {
		t.log.Error("Add DreamBooth job err:", zap.Error(err))
		return []byte(respBodySessionIsNotExist), ""
	}
	body, fileName, err := t.dreamBooth.GenerateImage(ctx, text)
	if errors.Is(err, context.Canceled) {
		return []byte(respErrBodyJobCanceled), ""
	}
	defer func() {
		if err = t.clientStates.ClientCancelDreamBoothJob(jobID, chatID); err != nil {
			t.log.Error("Cancel DreamBooth job err:", zap.Error(err))
		}
	}()
	if err != nil {
		t.log.Error("DreamBooth response err:", zap.Error(err))
		body = []byte(respErrBodyCommandDreamBooth(err))
	}
	return body, fileName
}

func (t *TBotOpenAI) processFusionBrain(text string, chatID int64) ([]byte, string) {
	ctx, cancel := context.WithTimeout(context.Background(), t.cfg.FusionBrain.Timeout)
	jobID := randIntByRange(minJobID, maxJobID)
	if err := t.clientStates.ClientAddFusionBrainJob(cancel, jobID, chatID); err != nil {
		t.log.Error("Add FusionBrain job err:", zap.Error(err))
		return []byte(respBodySessionIsNotExist), ""
	}
	body, fileName, err := t.fusionBrain.GenerateImage(ctx, text)
	if errors.Is(err, context.Canceled) {
		return []byte(respErrBodyJobCanceled), ""
	}
	defer func() {
		if err = t.clientStates.ClientCancelFusionBrainJob(jobID, chatID); err != nil {
			t.log.Error("Cancel FusionBrain job err:", zap.Error(err))
		}
	}()
	if err != nil {
		t.log.Error("FusionBrain response err:", zap.Error(err))
		body = []byte(respErrBodyFusionBrain)
	}
	return body, fileName
}

func (t *TBotOpenAI) writeStats(command, username, request, response string) {
	switch command {
	case commandChatGPT, commandOpenAIImage, commandOpenAIText, commandDreamBooth:
		loc, err := time.LoadLocation("Europe/Moscow")
		if err != nil {
			t.log.Error("Load location err:", zap.Error(err))
			return
		}
		t.stats.Write(statRow{
			ts:       time.Now().In(loc).Format(time.RFC3339),
			username: username,
			ai:       command,
			request:  request,
			response: prepareResponse(response),
		})
	}
}

func (t *TBotOpenAI) processBan(text string, _ int64) ([]byte, string) {
	_, ok := t.blacklist.LoadOrStore(text, struct{}{})
	if ok {
		return []byte(respErrBodyRequestBanUsernameAlreadyExist), ""
	}
	if err := t.writeBlacklistToFile(); err != nil {
		return []byte(respErrBodyRequestBan), ""
	}
	return []byte(respBodyRequestBan), ""
}

func (t *TBotOpenAI) processUnban(text string, _ int64) ([]byte, string) {
	_, ok := t.blacklist.LoadAndDelete(text)
	if !ok {
		return []byte(respErrBodyRequestUnbanUsernameIsNotExist), ""
	}
	if err := t.writeBlacklistToFile(); err != nil {
		return []byte(respErrBodyRequestUnban), ""
	}
	return []byte(respBodyRequestUnban), ""
}

func prepareResponse(response string) string {
	response = strings.ReplaceAll(response, "\n", "")
	return strings.ReplaceAll(response, "\r", "")
}

func randIntByRange(min, max int) int {
	return rand.Intn(max-min) + min
}

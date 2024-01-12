package tbotopenai

import (
	"context"
	"errors"
	"math/rand"
	"strconv"
	"time"

	"go.uber.org/zap"
)

const (
	minJobID = 100000
	maxJobID = 999999
)

const (
	labelChatGPT    = "ChatGPT"
	labelDreamBooth = "DreamBooth"
)

func (t *TBotOpenAI) processTask(text string, chatID int64) ([]byte, string) {
	command, err := t.clientStates.ClientCommand(chatID)
	if err != nil {
		t.log.Error("Get client command err:", zap.Error(err))
		return nil, ""
	}
	f, ok := t.taskByCmd[command]
	if !ok {
		return []byte(respBodyUndefinedGeneration), ""
	}
	return f(text, chatID)
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
	if err = t.clientStates.ClientCancelDreamBoothJob(jobID, chatID); err == nil {
		return respBodySuccessCancelJob(labelDreamBooth, jobID), ""
	}
	return respErrBodyJobIsNotExist(jobID), ""
}

func (t *TBotOpenAI) processChatGPT(text string, chatID int64) ([]byte, string) {
	if body := t.checkClientChatGPTJobs(chatID); body != "" {
		return []byte(body), ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(t.cfg.ChatGPT.Timeout)*time.Second)
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
	if body := t.checkClientOpenAIJobs(chatID); body != "" {
		return []byte(body), ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(t.cfg.OpenAI.Timeout)*time.Second)
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
	if body := t.checkClientOpenAIJobs(chatID); body != "" {
		return []byte(body), ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(t.cfg.OpenAI.Timeout)*time.Second)
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
	if body := t.checkClientDreamBoothJobs(chatID); body != "" {
		return []byte(body), ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(t.cfg.DreamBooth.Timeout)*time.Second)
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

func randIntByRange(min, max int) int {
	return rand.Intn(max-min) + min
}

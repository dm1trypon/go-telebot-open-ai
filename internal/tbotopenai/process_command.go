package tbotopenai

import "go.uber.org/zap"

func (t *TBotOpenAI) processCommand(command string, chatID int64) string {
	if command == "" {
		return ""
	}
	f, ok := t.clientStateByCmd[command]
	if !ok {
		return respBodyUndefinedCommand
	}
	return f(command, chatID)
}

func (t *TBotOpenAI) commandHelp(_ string, _ int64) string {
	return respBodyCommandHelp
}

func (t *TBotOpenAI) commandDreamBoothExample(_ string, _ int64) string {
	return respBodyCommandDreamBoothExample
}

func (t *TBotOpenAI) commandStart(_ string, chatID int64) string {
	if err := t.clientStates.AddClient(chatID); err != nil {
		t.log.Error("Add client err:", zap.Error(err))
		return respBodySessionIsAlreadyExist
	}
	return respBodySessionCreated
}

func (t *TBotOpenAI) commandStop(_ string, chatID int64) string {
	if err := t.clientStates.ClientCancelJobs(chatID); err != nil {
		t.log.Error("Cancel client jobs err:", zap.Error(err))
		return respBodySessionIsNotExist
	}
	if err := t.clientStates.DeleteClient(chatID); err != nil {
		t.log.Error("Delete clientState err:", zap.Error(err))
		return respBodySessionIsAlreadyExist
	}
	return respBodySessionRemoved
}

func (t *TBotOpenAI) commandDreamBooth(command string, chatID int64) string {
	if err := t.clientStates.UpdateClientCommand(chatID, command); err != nil {
		t.log.Error("Update client command err:", zap.Error(err))
		return respBodySessionIsNotExist
	}
	return respBodyCommandDreamBooth
}

func (t *TBotOpenAI) commandChatGPT(command string, chatID int64) string {
	if err := t.clientStates.UpdateClientCommand(chatID, command); err != nil {
		t.log.Error("Update client command err:", zap.Error(err))
		return respBodySessionIsNotExist
	}
	return respBodyCommandChatGPT
}

func (t *TBotOpenAI) commandOpenAIText(command string, chatID int64) string {
	if err := t.clientStates.UpdateClientCommand(chatID, command); err != nil {
		t.log.Error("Update client command err:", zap.Error(err))
		return respBodySessionIsNotExist
	}
	return respBodyCommandOpenAIText
}

func (t *TBotOpenAI) commandOpenAIImage(command string, chatID int64) string {
	if err := t.clientStates.UpdateClientCommand(chatID, command); err != nil {
		t.log.Error("Update client command err:", zap.Error(err))
		return respBodySessionIsNotExist
	}
	return respBodyCommandOpenAIImage
}

func (t *TBotOpenAI) commandCancelJob(command string, chatID int64) string {
	if err := t.clientStates.UpdateClientCommand(chatID, command); err != nil {
		t.log.Error("Update client command err:", zap.Error(err))
		return respBodySessionIsNotExist
	}
	return respBodyInputJobID
}

func (t *TBotOpenAI) commandListJobs(_ string, chatID int64) string {
	textJobIDs, err := t.clientStates.ClientChatGPTJobs(chatID)
	if err != nil {
		t.log.Error("Get ChatGPT jobs err:", zap.Error(err))
		return respBodySessionIsNotExist
	}
	imgJobIDs, err := t.clientStates.ClientDreamBoothJobs(chatID)
	if err != nil {
		t.log.Error("Get DreamBooth jobs err:", zap.Error(err))
		return respBodySessionIsNotExist
	}
	openAIIDs, err := t.clientStates.ClientOpenAIJobs(chatID)
	if err != nil {
		t.log.Error("Get OpenAI jobs err:", zap.Error(err))
		return respBodySessionIsNotExist
	}
	return respBodyListJobs(textJobIDs, imgJobIDs, openAIIDs)
}

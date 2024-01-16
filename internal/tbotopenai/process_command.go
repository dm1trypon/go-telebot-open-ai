package tbotopenai

import "go.uber.org/zap"

func (t *TBotOpenAI) processCommand(command, username string, chatID int64) string {
	if command == "" {
		return ""
	}
	if !t.checkPermissions(command, username) {
		return respBodyUndefinedCommand
	}
	f, ok := t.clientStateByCmd[command]
	if !ok {
		return respBodyUndefinedCommand
	}
	return f(command, username, chatID)
}

func (t *TBotOpenAI) commandHelp(_, _ string, _ int64) string {
	return respBodyCommandHelp
}

func (t *TBotOpenAI) commandDreamBoothExample(_, _ string, _ int64) string {
	return respBodyCommandDreamBoothExample
}

func (t *TBotOpenAI) commandStart(_, username string, chatID int64) string {
	if err := t.clientStates.AddClient(chatID, username); err != nil {
		t.log.Error("Add client err:", zap.Error(err))
		return respBodySessionIsAlreadyExist
	}
	return respBodySessionCreated
}

func (t *TBotOpenAI) commandStop(_, _ string, chatID int64) string {
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

func (t *TBotOpenAI) commandDreamBooth(command, _ string, chatID int64) string {
	if err := t.clientStates.UpdateClientCommand(chatID, command); err != nil {
		t.log.Error("Update client command err:", zap.Error(err))
		return respBodySessionIsNotExist
	}
	return respBodyCommandDreamBooth
}

func (t *TBotOpenAI) commandChatGPT(command, _ string, chatID int64) string {
	if err := t.clientStates.UpdateClientCommand(chatID, command); err != nil {
		t.log.Error("Update client command err:", zap.Error(err))
		return respBodySessionIsNotExist
	}
	return respBodyCommandChatGPT
}

func (t *TBotOpenAI) commandOpenAIText(command, _ string, chatID int64) string {
	if err := t.clientStates.UpdateClientCommand(chatID, command); err != nil {
		t.log.Error("Update client command err:", zap.Error(err))
		return respBodySessionIsNotExist
	}
	return respBodyCommandOpenAIText
}

func (t *TBotOpenAI) commandOpenAIImage(command, _ string, chatID int64) string {
	if err := t.clientStates.UpdateClientCommand(chatID, command); err != nil {
		t.log.Error("Update client command err:", zap.Error(err))
		return respBodySessionIsNotExist
	}
	return respBodyCommandOpenAIImage
}

func (t *TBotOpenAI) commandCancelJob(command, _ string, chatID int64) string {
	if err := t.clientStates.UpdateClientCommand(chatID, command); err != nil {
		t.log.Error("Update client command err:", zap.Error(err))
		return respBodySessionIsNotExist
	}
	return respBodyInputJobID
}

func (t *TBotOpenAI) commandListJobs(_, _ string, chatID int64) string {
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

func (t *TBotOpenAI) commandHistory(_, username string, _ int64) string {
	if _, ok := t.userRoles[roleAdmin][username]; !ok {
		return respBodyUndefinedCommand
	}
	return respBodyHistoryCommand
}

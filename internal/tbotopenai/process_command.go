package tbotopenai

import "go.uber.org/zap"

func (t *TBotOpenAI) processCommand(command, username string, chatID int64) (string, []byte) {
	if command == "" {
		return "", nil
	}
	if !t.checkPermissions(command, username) {
		return respBodyUndefinedCommand, nil
	}
	f, ok := t.clientStateByCmd[command]
	if !ok {
		return respBodyUndefinedCommand, nil
	}
	return f(command, username, chatID)
}

func (t *TBotOpenAI) commandHelp(_, username string, _ int64) (string, []byte) {
	curRole := t.getRole(username)
	if curRole == "" {
		return respBodyUndefinedCommand, nil
	}
	return respBodyCommandHelp(curRole), nil
}

func (t *TBotOpenAI) commandDreamBoothExample(_, _ string, _ int64) (string, []byte) {
	return respBodyCommandDreamBoothExample, nil
}

func (t *TBotOpenAI) commandStart(_, username string, chatID int64) (string, []byte) {
	if err := t.clientStates.AddClient(chatID, username); err != nil {
		t.log.Error("Add client err:", zap.Error(err))
		return respBodySessionIsAlreadyExist, nil
	}
	return respBodySessionCreated, nil
}

func (t *TBotOpenAI) commandStop(_, _ string, chatID int64) (string, []byte) {
	if err := t.clientStates.ClientCancelJobs(chatID); err != nil {
		t.log.Error("Cancel client jobs err:", zap.Error(err))
		return respBodySessionIsNotExist, nil
	}
	if err := t.clientStates.DeleteClient(chatID); err != nil {
		t.log.Error("Delete clientState err:", zap.Error(err))
		return respBodySessionIsAlreadyExist, nil
	}
	return respBodySessionRemoved, nil
}

func (t *TBotOpenAI) commandDreamBooth(command, _ string, chatID int64) (string, []byte) {
	if err := t.clientStates.UpdateClientCommand(chatID, command); err != nil {
		t.log.Error("Update client command err:", zap.Error(err))
		return respBodySessionIsNotExist, nil
	}
	return respBodyCommandDreamBooth, nil
}

func (t *TBotOpenAI) commandChatGPT(command, _ string, chatID int64) (string, []byte) {
	if err := t.clientStates.UpdateClientCommand(chatID, command); err != nil {
		t.log.Error("Update client command err:", zap.Error(err))
		return respBodySessionIsNotExist, nil
	}
	return respBodyCommandChatGPT, nil
}

func (t *TBotOpenAI) commandOpenAIText(command, _ string, chatID int64) (string, []byte) {
	if err := t.clientStates.UpdateClientCommand(chatID, command); err != nil {
		t.log.Error("Update client command err:", zap.Error(err))
		return respBodySessionIsNotExist, nil
	}
	return respBodyCommandOpenAIText, nil
}

func (t *TBotOpenAI) commandOpenAIImage(command, _ string, chatID int64) (string, []byte) {
	if err := t.clientStates.UpdateClientCommand(chatID, command); err != nil {
		t.log.Error("Update client command err:", zap.Error(err))
		return respBodySessionIsNotExist, nil
	}
	return respBodyCommandOpenAIImage, nil
}

func (t *TBotOpenAI) commandCancelJob(command, _ string, chatID int64) (string, []byte) {
	if err := t.clientStates.UpdateClientCommand(chatID, command); err != nil {
		t.log.Error("Update client command err:", zap.Error(err))
		return respBodySessionIsNotExist, nil
	}
	return respBodyInputJobID, nil
}

func (t *TBotOpenAI) commandListJobs(_, _ string, chatID int64) (string, []byte) {
	textJobIDs, err := t.clientStates.ClientChatGPTJobs(chatID)
	if err != nil {
		t.log.Error("Get ChatGPT jobs err:", zap.Error(err))
		return respBodySessionIsNotExist, nil
	}
	imgJobIDs, err := t.clientStates.ClientDreamBoothJobs(chatID)
	if err != nil {
		t.log.Error("Get DreamBooth jobs err:", zap.Error(err))
		return respBodySessionIsNotExist, nil
	}
	openAIIDs, err := t.clientStates.ClientOpenAIJobs(chatID)
	if err != nil {
		t.log.Error("Get OpenAI jobs err:", zap.Error(err))
		return respBodySessionIsNotExist, nil
	}
	return respBodyListJobs(textJobIDs, imgJobIDs, openAIIDs), nil
}

func (t *TBotOpenAI) commandHistory(_, _ string, _ int64) (string, []byte) {
	return respBodyHistoryCommand, t.stats.Bytes()
}

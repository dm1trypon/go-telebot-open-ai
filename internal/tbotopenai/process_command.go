package tbotopenai

import (
	"bufio"
	"bytes"
	"os"

	"go.uber.org/zap"
)

const fileNameLogs = "logs.log"

type commandResponse struct {
	text     string
	fileName string
	fileBody []byte
}

func (t *TBotOpenAI) processCommand(command, username string, chatID int64) *commandResponse {
	if command == "" {
		return nil
	}
	if !t.checkPermissions(command, username) {
		return &commandResponse{
			text: respBodyUndefinedCommand,
		}
	}
	f, ok := t.clientStateByCmd[command]
	if !ok {
		return &commandResponse{
			text: respBodyUndefinedCommand,
		}
	}
	return f(command, username, chatID)
}

func (t *TBotOpenAI) commandHelp(_, username string, _ int64) *commandResponse {
	curRole := t.getRole(username)
	if curRole == "" {
		return &commandResponse{
			text: respBodyUndefinedCommand,
		}
	}
	return &commandResponse{
		text: respBodyCommandHelp(curRole),
	}
}

func (t *TBotOpenAI) commandDreamBoothExample(_, _ string, _ int64) *commandResponse {
	return &commandResponse{
		text: respBodyCommandDreamBoothExample,
	}
}

func (t *TBotOpenAI) commandStart(_, username string, chatID int64) *commandResponse {
	if err := t.clientStates.AddClient(chatID, username); err != nil {
		t.log.Error("Add client err:", zap.Error(err))
		return &commandResponse{
			text: respBodySessionIsAlreadyExist,
		}
	}
	return &commandResponse{
		text: respBodySessionCreated,
	}
}

func (t *TBotOpenAI) commandStop(_, _ string, chatID int64) *commandResponse {
	if err := t.clientStates.ClientCancelJobs(chatID); err != nil {
		t.log.Error("Cancel client jobs err:", zap.Error(err))
		return &commandResponse{
			text: respBodySessionIsNotExist,
		}
	}
	if err := t.clientStates.DeleteClient(chatID); err != nil {
		t.log.Error("Delete clientState err:", zap.Error(err))
		return &commandResponse{
			text: respBodySessionIsNotExist,
		}
	}
	return &commandResponse{
		text: respBodySessionRemoved,
	}
}

func (t *TBotOpenAI) commandDreamBooth(command, _ string, chatID int64) *commandResponse {
	if err := t.clientStates.UpdateClientCommand(chatID, command); err != nil {
		t.log.Error("Update client command err:", zap.Error(err))
		return &commandResponse{
			text: respBodySessionIsNotExist,
		}
	}
	return &commandResponse{
		text: respBodyCommandDreamBooth,
	}
}

func (t *TBotOpenAI) commandChatGPT(command, _ string, chatID int64) *commandResponse {
	if err := t.clientStates.UpdateClientCommand(chatID, command); err != nil {
		t.log.Error("Update client command err:", zap.Error(err))
		return &commandResponse{
			text: respBodySessionIsNotExist,
		}
	}
	return &commandResponse{
		text: respBodyCommandChatGPT,
	}
}

func (t *TBotOpenAI) commandOpenAIText(command, _ string, chatID int64) *commandResponse {
	if err := t.clientStates.UpdateClientCommand(chatID, command); err != nil {
		t.log.Error("Update client command err:", zap.Error(err))
		return &commandResponse{
			text: respBodySessionIsNotExist,
		}
	}
	return &commandResponse{
		text: respBodyCommandOpenAIText,
	}
}

func (t *TBotOpenAI) commandOpenAIImage(command, _ string, chatID int64) *commandResponse {
	if err := t.clientStates.UpdateClientCommand(chatID, command); err != nil {
		t.log.Error("Update client command err:", zap.Error(err))
		return &commandResponse{
			text: respBodySessionIsNotExist,
		}
	}
	return &commandResponse{
		text: respBodyCommandOpenAIImage,
	}
}

func (t *TBotOpenAI) commandCancelJob(command, _ string, chatID int64) *commandResponse {
	if err := t.clientStates.UpdateClientCommand(chatID, command); err != nil {
		t.log.Error("Update client command err:", zap.Error(err))
		return &commandResponse{
			text: respBodySessionIsNotExist,
		}
	}
	return &commandResponse{
		text: respBodyInputJobID,
	}
}

func (t *TBotOpenAI) commandListJobs(_, _ string, chatID int64) *commandResponse {
	textJobIDs, err := t.clientStates.ClientChatGPTJobs(chatID)
	if err != nil {
		t.log.Error("Get ChatGPT jobs err:", zap.Error(err))
		return &commandResponse{
			text: respBodySessionIsNotExist,
		}
	}
	imgJobIDs, err := t.clientStates.ClientDreamBoothJobs(chatID)
	if err != nil {
		t.log.Error("Get DreamBooth jobs err:", zap.Error(err))
		return &commandResponse{
			text: respBodySessionIsNotExist,
		}
	}
	openAIIDs, err := t.clientStates.ClientOpenAIJobs(chatID)
	if err != nil {
		t.log.Error("Get OpenAI jobs err:", zap.Error(err))
		return &commandResponse{
			text: respBodySessionIsNotExist,
		}
	}
	return &commandResponse{
		text: respBodyListJobs(textJobIDs, imgJobIDs, openAIIDs),
	}
}

func (t *TBotOpenAI) commandStats(_, _ string, _ int64) *commandResponse {
	statsBody := t.stats.Bytes()
	if len(statsBody) == 0 {
		return &commandResponse{
			text: respBodyStatsCommand,
		}
	}
	return &commandResponse{
		fileName: fileNameStats,
		fileBody: t.stats.Bytes(),
	}
}

func (t *TBotOpenAI) commandLogs(_, _ string, _ int64) *commandResponse {
	if len(t.cfg.Logger.OutputPaths) == 0 {
		t.log.Error("Empty output paths for logs")
		return &commandResponse{
			text: respErrBodyGetLogs,
		}
	}
	file, err := os.Open(t.cfg.Logger.OutputPaths[0])
	if err != nil {
		t.log.Error("Reading log's file err:", zap.Error(err))
		return &commandResponse{
			text: respErrBodyGetLogs,
		}
	}
	defer func() {
		if err = file.Close(); err != nil {
			t.log.Error("Close log's file err:", zap.Error(err))
		}
	}()
	scanner := bufio.NewScanner(file)
	rows := make([][]byte, 0, t.cfg.MaxLogRows)
	for scanner.Scan() {
		row := scanner.Bytes()
		rows = append(rows, row)
		if len(rows) > t.cfg.MaxLogRows {
			rows = rows[1:]
		}
	}
	if err = scanner.Err(); err != nil {
		t.log.Error("Scanner log's file err:", zap.Error(err))
		return &commandResponse{
			text: respErrBodyGetLogs,
		}
	}
	return &commandResponse{
		fileName: fileNameLogs,
		fileBody: bytes.Join(rows, []byte("\n\r")),
	}
}

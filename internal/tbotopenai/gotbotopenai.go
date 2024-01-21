package tbotopenai

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

const (
	commandStart              = "start"
	commandStop               = "stop"
	commandChatGPT            = "chatGPT"
	commandOpenAIText         = "openAIText"
	commandOpenAIImage        = "openAIImage"
	commandDreamBooth         = "dreamBooth"
	commandImageCustomExample = "dreamBoothExample"
	commandHelp               = "help"
	commandCancelJob          = "cancelJob"
	commandListJobs           = "listJobs"
	commandStats              = "stats"
)

const (
	roleAdmin = "admin"
	roleUser  = "user"
)

type AI interface {
	GenerateText(ctx context.Context, prompt string) ([]byte, error)
	GenerateImage(ctx context.Context, prompt string) ([]byte, string, error)
}

type TBotOpenAI struct {
	cfg              *Config
	telegram         Messenger
	dreamBooth       AI
	openAI           AI
	chatGPTBot       AI
	clientStates     clientStateByChatID
	stats            *Stats
	log              *zap.Logger
	msgChan          chan *message
	queueTaskChan    chan *message
	userRoles        map[string]map[string]struct{}
	permissions      map[string]map[string]struct{}
	taskByCmd        map[string]func(text string, chatID int64) ([]byte, string)
	clientStateByCmd map[string]func(command, username string, chatID int64) (string, []byte)
}

func NewTBotOpenAI(cfg *Config, log *zap.Logger) (*TBotOpenAI, error) {
	msgChan := make(chan *message, cfg.LenMessageChan)
	queueTaskChan := make(chan *message, cfg.LenQueueTaskChan)
	telegram, err := NewTelegram(&cfg.Telegram, log, msgChan)
	if err != nil {
		return nil, err
	}
	g := &TBotOpenAI{
		cfg:           cfg,
		telegram:      telegram,
		dreamBooth:    NewDreamBoothAPI(log, &cfg.DreamBooth),
		openAI:        NewOpenAI(&cfg.OpenAI),
		chatGPTBot:    NewChatGPTBot(),
		clientStates:  clientStateByChatID{value: make(map[int64]*clientState)},
		stats:         NewStats(log, cfg.Stats.Interval, cfg.Stats.Filepath),
		log:           log,
		msgChan:       msgChan,
		queueTaskChan: queueTaskChan,
	}
	g.setUserRoles(&cfg.Roles)
	g.setPermissions(&cfg.Permissions)
	g.taskByCmd = make(map[string]func(text string, chatID int64) (body []byte, fileName string), 5)
	g.taskByCmd[commandChatGPT] = g.processChatGPT
	g.taskByCmd[commandDreamBooth] = g.processDreamBooth
	g.taskByCmd[commandCancelJob] = g.processCancelJob
	g.taskByCmd[commandOpenAIText] = g.processOpenAIText
	g.taskByCmd[commandOpenAIImage] = g.processOpenAIImage
	g.clientStateByCmd = make(map[string]func(command, username string, chatID int64) (string, []byte), 11)
	g.clientStateByCmd[commandHelp] = g.commandHelp
	g.clientStateByCmd[commandImageCustomExample] = g.commandDreamBoothExample
	g.clientStateByCmd[commandStart] = g.commandStart
	g.clientStateByCmd[commandStop] = g.commandStop
	g.clientStateByCmd[commandChatGPT] = g.commandChatGPT
	g.clientStateByCmd[commandDreamBooth] = g.commandDreamBooth
	g.clientStateByCmd[commandOpenAIText] = g.commandOpenAIText
	g.clientStateByCmd[commandOpenAIImage] = g.commandOpenAIImage
	g.clientStateByCmd[commandCancelJob] = g.commandCancelJob
	g.clientStateByCmd[commandListJobs] = g.commandListJobs
	g.clientStateByCmd[commandStats] = g.commandStats
	return g, nil
}

func (t *TBotOpenAI) Run() {
	var wg sync.WaitGroup
	if err := t.stats.Run(&wg); err != nil {
		t.log.Error("Running Stats err", zap.Error(err))
		return
	}
	t.initQueueTaskWorkers(&wg)
	wg.Add(1)
	go t.initProcessMessagesWorker(&wg)
	t.telegram.Run()
	wg.Wait()
}

func (t *TBotOpenAI) Stop() {
	t.telegram.Stop()
	t.stats.Stop()
	close(t.msgChan)
	close(t.queueTaskChan)
}

func (t *TBotOpenAI) initProcessMessagesWorker(wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			t.log.Error("Recovered panic err:", zap.Any("panic", r))
		}
	}()
	defer wg.Done()
	for {
		select {
		case msg, ok := <-t.msgChan:
			if !ok {
				return
			}
			t.log.Debug("Received message",
				zap.String("user", msg.username),
				zap.String("body", msg.text),
				zap.String("command", msg.command))
			respBody := t.checkChanMessagesBuffer()
			if respBody != "" {
				if err := t.telegram.ReplyText(msg.messageID, msg.chatID, respBody); err != nil {
					t.log.Error("Reply message error:", zap.Error(err))
				}
				continue
			}
			var fileBody []byte
			respBody, fileBody = t.processCommand(msg.command, msg.username, msg.chatID)
			if respBody != "" {
				if err := t.telegram.ReplyText(msg.messageID, msg.chatID, respBody); err != nil {
					t.log.Error("Reply message error:", zap.Error(err))
				}
				continue
			} else if fileBody != nil {
				if err := t.telegram.ReplyFile(msg.messageID, msg.chatID, fileBody, fileNameStats); err != nil {
					t.log.Error("Reply message error:", zap.Error(err))
				}
				continue
			}
			if msg.text == "" {
				continue
			}
			respBody = t.checkJobsLimit(msg.chatID)
			if respBody != "" {
				if err := t.telegram.ReplyText(msg.messageID, msg.chatID, respBody); err != nil {
					t.log.Error("Reply message error:", zap.Error(err))
				}
				continue
			}
			if err := t.telegram.ReplyText(msg.messageID, msg.chatID, respBodyRequestAddedToQueue); err != nil {
				t.log.Error("Reply message error:", zap.Error(err))
			}
			t.queueTaskChan <- msg
		}
	}
}

func (t *TBotOpenAI) checkJobsLimit(chatID int64) string {
	command, err := t.clientStates.ClientCommand(chatID)
	if err != nil {
		t.log.Error("Get client command err:", zap.Error(err))
		return respBodySessionIsNotExist
	}
	switch command {
	case commandChatGPT:
		if body := t.checkClientChatGPTJobs(chatID); body != "" {
			return body
		}
	case commandOpenAIText, commandOpenAIImage:
		if body := t.checkClientOpenAIJobs(chatID); body != "" {
			return body
		}
	case commandDreamBooth:
		if body := t.checkClientDreamBoothJobs(chatID); body != "" {
			return body
		}
	}
	return ""
}

func (t *TBotOpenAI) initQueueTaskWorkers(wg *sync.WaitGroup) {
	wg.Add(t.cfg.QueueMessageWorkers)
	for i := 0; i < t.cfg.QueueMessageWorkers; i++ {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					t.log.Error("Recovered panic err:", zap.Any("panic", r))
				}
			}()
			defer wg.Done()
			for {
				select {
				case msg, ok := <-t.queueTaskChan:
					if !ok {
						return
					}
					t.processQueueTask(msg.text, msg.messageID, msg.chatID)
				}
			}
		}()
	}
}

func (t *TBotOpenAI) processQueueTask(text string, messageID int, chatID int64) {
	var err error
	body, fileName := t.processTask(text, chatID)
	if fileName != "" {
		err = t.telegram.ReplyFile(messageID, chatID, body, fileName)
	} else {
		err = t.telegram.ReplyText(messageID, chatID, string(body))
	}
	if err != nil {
		t.log.Error("Reply to client err:", zap.Error(err))
	}
}

func (t *TBotOpenAI) checkClientChatGPTJobs(chatID int64) string {
	jobs, err := t.clientStates.ClientLenChatGPTJobs(chatID)
	if err != nil {
		t.log.Error("Get ChatGPT jobs err:", zap.Error(err))
		return respBodySessionIsNotExist
	}
	if jobs >= t.cfg.MaxClientOpenAIJobs {
		return respErrBodyLimitJobs
	}
	return ""
}

func (t *TBotOpenAI) checkClientDreamBoothJobs(chatID int64) string {
	jobs, err := t.clientStates.ClientLenDreamBoothJobs(chatID)
	if err != nil {
		t.log.Error("Get DreamBooth jobs err:", zap.Error(err))
		return respBodySessionIsNotExist
	}
	if jobs >= t.cfg.MaxClientDreamBoothJobs {
		return respErrBodyLimitJobs
	}
	return ""
}

func (t *TBotOpenAI) checkClientOpenAIJobs(chatID int64) string {
	jobs, err := t.clientStates.ClientLenOpenAIJobs(chatID)
	if err != nil {
		t.log.Error("Get OpenAI jobs err:", zap.Error(err))
		return respBodySessionIsNotExist
	}
	if jobs >= t.cfg.MaxClientOpenAIJobs {
		return respErrBodyLimitJobs
	}
	return ""
}

func (t *TBotOpenAI) checkChanMessagesBuffer() string {
	if len(t.queueTaskChan) >= t.cfg.LenMessageChan {
		return respErrBodyLimitMessages
	}
	return ""
}

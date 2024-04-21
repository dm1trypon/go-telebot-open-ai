package tbotopenai

import (
	"bytes"
	"context"
	"os"
	"strings"
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
	commandDreamBoothExample  = "dreamBoothExample"
	commandFusionBrain        = "fusionBrain"
	commandFusionBrainExample = "fusionBrainExample"
	commandHelp               = "help"
	commandCancelJob          = "cancelJob"
	commandListJobs           = "listJobs"
	commandStats              = "stats"
	commandLogs               = "logs"
	commandBan                = "ban"
	commandUnban              = "unban"
	commandBlacklist          = "blacklist"
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
	fusionBrain      AI
	clientStates     clientStateByChatID
	stats            *Stats
	log              *zap.Logger
	msgChan          chan *message
	queueTaskChan    chan *message
	userRoles        sync.Map
	permissions      sync.Map
	taskByCmd        sync.Map
	clientStateByCmd sync.Map
	blacklist        sync.Map
}

func NewTBotOpenAI(cfg *Config, log *zap.Logger) (*TBotOpenAI, error) {
	msgChan := make(chan *message, cfg.LenMessageChan)
	queueTaskChan := make(chan *message, cfg.LenQueueTaskChan)
	telegram, err := NewTelegram(&cfg.Telegram, log, msgChan)
	if err != nil {
		return nil, err
	}
	t := &TBotOpenAI{
		cfg:           cfg,
		telegram:      telegram,
		dreamBooth:    NewDreamBoothAPI(log, &cfg.DreamBooth),
		openAI:        NewOpenAI(&cfg.OpenAI),
		chatGPTBot:    NewChatGPTBot(),
		fusionBrain:   NewFusionBrainAPI(log, &cfg.FusionBrain),
		clientStates:  clientStateByChatID{value: make(map[int64]*clientState)},
		stats:         NewStats(log, cfg.Stats.Interval, cfg.Stats.Filepath),
		log:           log,
		msgChan:       msgChan,
		queueTaskChan: queueTaskChan,
	}
	t.setUserRoles(&cfg.Roles)
	t.setPermissions(&cfg.Permissions)
	t.taskByCmd.Store(commandChatGPT, t.processChatGPT)
	t.taskByCmd.Store(commandDreamBooth, t.processDreamBooth)
	t.taskByCmd.Store(commandCancelJob, t.processCancelJob)
	t.taskByCmd.Store(commandOpenAIText, t.processOpenAIText)
	t.taskByCmd.Store(commandOpenAIImage, t.processOpenAIImage)
	t.taskByCmd.Store(commandFusionBrain, t.processFusionBrain)
	t.taskByCmd.Store(commandBan, t.processBan)
	t.taskByCmd.Store(commandUnban, t.processUnban)
	t.clientStateByCmd.Store(commandHelp, t.commandHelp)
	t.clientStateByCmd.Store(commandDreamBoothExample, t.commandDreamBoothExample)
	t.clientStateByCmd.Store(commandFusionBrainExample, t.commandFusionBrainExample)
	t.clientStateByCmd.Store(commandStart, t.commandStart)
	t.clientStateByCmd.Store(commandStop, t.commandStop)
	t.clientStateByCmd.Store(commandChatGPT, t.commandChatGPT)
	t.clientStateByCmd.Store(commandDreamBooth, t.commandDreamBooth)
	t.clientStateByCmd.Store(commandOpenAIText, t.commandOpenAIText)
	t.clientStateByCmd.Store(commandOpenAIImage, t.commandOpenAIImage)
	t.clientStateByCmd.Store(commandFusionBrain, t.commandFusionBrain)
	t.clientStateByCmd.Store(commandCancelJob, t.commandCancelJob)
	t.clientStateByCmd.Store(commandListJobs, t.commandListJobs)
	t.clientStateByCmd.Store(commandStats, t.commandStats)
	t.clientStateByCmd.Store(commandLogs, t.commandLogs)
	t.clientStateByCmd.Store(commandBan, t.commandBan)
	t.clientStateByCmd.Store(commandUnban, t.commandUnban)
	t.clientStateByCmd.Store(commandBlacklist, t.commandBlacklist)
	if err = t.storeBlacklist(); err != nil {
		return nil, err
	}
	return t, nil
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
			if t.isBanned(msg.username) {
				if err := t.telegram.ReplyText(msg.messageID, msg.chatID, respBodyAccessDenied); err != nil {
					t.log.Error("Reply message error:", zap.Error(err))
				}
				continue
			}
			respBody := t.checkChanMessagesBuffer()
			if respBody != "" {
				if err := t.telegram.ReplyText(msg.messageID, msg.chatID, respBody); err != nil {
					t.log.Error("Reply message error:", zap.Error(err))
				}
				continue
			}
			resp := t.processCommand(msg.command, msg.username, msg.chatID)
			if resp != nil && resp.text != "" {
				if err := t.telegram.ReplyText(msg.messageID, msg.chatID, resp.text); err != nil {
					t.log.Error("Reply message error:", zap.Error(err))
				}
				continue
			} else if resp != nil && resp.fileBody != nil {
				if err := t.telegram.ReplyFile(msg.messageID, msg.chatID, resp.fileBody, resp.fileName); err != nil {
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

func (t *TBotOpenAI) storeBlacklist() error {
	f, err := os.OpenFile(t.cfg.PathBlackList, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer func() {
		if err = f.Close(); err != nil {
			t.log.Error("Open blacklist's file err", zap.Error(err))
		}
	}()
	fStat, err := f.Stat()
	if err != nil {
		t.log.Error("Read stat blacklist's file err", zap.Error(err))
		return err
	}
	body := make([]byte, fStat.Size())
	if _, err = f.Read(body); err != nil {
		t.log.Error("Read blacklist's file err", zap.Error(err))
		return err
	}
	if len(body) == 0 {
		return nil
	}
	strBody := string(body)
	strBody = strings.ReplaceAll(strBody, "\r", "")
	rows := strings.Split(strBody, "\n")
	for _, row := range rows {
		t.blacklist.Store(row, struct{}{})
	}
	return nil
}

func (t *TBotOpenAI) writeBlacklistToFile() error {
	var b bytes.Buffer
	t.blacklist.Range(func(k, v any) bool {
		username, ok := k.(string)
		if !ok {
			return false
		}
		b.WriteString(username + "\n")
		return true
	})
	err := os.WriteFile(t.cfg.PathBlackList, b.Bytes(), 0644)
	if err != nil {
		t.log.Error("Write blacklist's file err", zap.Error(err))
	}
	return err
}

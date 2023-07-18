package gotbotopenai

import (
	"bytes"
	"context"
	"errors"
	"sync"

	"go.uber.org/zap"
)

const (
	commandStart              = "start"
	commandStop               = "stop"
	commandText               = "text"
	commandImageSize256x256   = "image256x256"
	commandImageSize512x512   = "image512x512 "
	commandImageSize1024x1024 = "image1024x1024"
	commandImageCustom        = "imageCustom"
	commandImageCustomExample = "imageCustomExample"
	commandHelp               = "help"

	resolution256x256   = "256x256"
	resolution512x512   = "512x512"
	resolution1024x1024 = "1024x1024"
)

var respBodyByCmd = map[string]string{
	commandStart:              respBodySessionCreated,
	commandStop:               respBodySessionRemoved,
	commandText:               respBodyCommandText,
	commandImageSize256x256:   respBodyCommandImage(resolution256x256),
	commandImageSize512x512:   respBodyCommandImage(resolution512x512),
	commandImageSize1024x1024: respBodyCommandImage(resolution1024x1024),
	commandImageCustom:        respBodyCommandImageCustom,
	commandImageCustomExample: respBodyCommandImageCustomExample,
	commandHelp:               respBodyCommandHelp,
}

var respErrBodyByCmd = map[string]string{
	commandText:               respErrBodyCommandText,
	commandImageSize256x256:   respErrBodyCommandImage,
	commandImageSize512x512:   respErrBodyCommandImage,
	commandImageSize1024x1024: respErrBodyCommandImage,
}

var resolutionByImgCommand = map[string]string{
	commandImageSize256x256:   resolution256x256,
	commandImageSize512x512:   resolution512x512,
	commandImageSize1024x1024: resolution1024x1024,
}

type tClient struct {
	curCmd  string
	curJobs int
}

type tClientByChatID struct {
	value map[int64]*tClient
	mutex sync.RWMutex
}

func (t *tClientByChatID) SetClientCurrentCommand(chatID int64, command string) {
	defer t.mutex.Unlock()
	t.mutex.Lock()
	tc, ok := t.value[chatID]
	if !ok || tc == nil {
		t.value[chatID] = &tClient{command, 0}
		return
	}
	tc.curCmd = command
}

func (t *tClientByChatID) IncrementClientCurrentJobs(chatID int64) error {
	defer t.mutex.Unlock()
	t.mutex.Lock()
	tc, ok := t.value[chatID]
	if !ok || tc == nil {
		return errors.New("client with current chatID does not exist")
	}
	tc.curJobs++
	return nil
}

func (t *tClientByChatID) DecrementClientCurrentJobs(chatID int64) error {
	defer t.mutex.Unlock()
	t.mutex.Lock()
	tc, ok := t.value[chatID]
	if !ok || tc == nil {
		return errors.New("client with current chatID does not exist")
	}
	if tc.curJobs == 0 {
		return errors.New("current jobs can not less than 0")
	}
	tc.curJobs--
	return nil
}

func (t *tClientByChatID) ClientCurrentJobs(chatID int64) (int, error) {
	defer t.mutex.RUnlock()
	t.mutex.RLock()
	tc, ok := t.value[chatID]
	if !ok || tc == nil {
		return -1, errors.New("client with current chatID does not exist")
	}
	return tc.curJobs, nil
}

func (t *tClientByChatID) ClientCurrentCommand(chatID int64) (string, error) {
	defer t.mutex.RUnlock()
	t.mutex.RLock()
	tc, ok := t.value[chatID]
	if !ok || tc == nil {
		return "", errors.New("client with current chatID does not exist")
	}
	return tc.curCmd, nil
}

func (t *tClientByChatID) AddClient(chatID int64) {
	defer t.mutex.Unlock()
	t.mutex.Lock()
	t.value[chatID] = new(tClient)
}

func (t *tClientByChatID) DeleteClient(chatID int64) {
	defer t.mutex.Unlock()
	t.mutex.Lock()
	delete(t.value, chatID)
}

type GoTBotOpenAI struct {
	cfg           *Config
	botClient     BotClient
	chatGPT       *ChatGPT
	dreamBoothAPI *DreamBoothAPI
	tClients      tClientByChatID
	log           *zap.Logger
	msgChan       chan *message
	quitChan      chan<- struct{}
}

func NewGoTBotOpenAI(cfg *Config, log *zap.Logger) (*GoTBotOpenAI, error) {
	msgChan := make(chan *message, cfg.LenMessageChan)
	quitChan := make(chan struct{}, 1)
	telegram, err := NewTelegram(&cfg.Telegram, log, msgChan, quitChan)
	if err != nil {
		return nil, err
	}
	return &GoTBotOpenAI{
		cfg:           cfg,
		botClient:     telegram,
		chatGPT:       NewChatGPT(&cfg.ChatGPT),
		dreamBoothAPI: NewDreamBoothAPI(log, &cfg.DreamBooth),
		tClients:      tClientByChatID{value: make(map[int64]*tClient)},
		log:           log,
		msgChan:       msgChan,
		quitChan:      quitChan,
	}, nil
}

func (g *GoTBotOpenAI) Run() {
	var wg sync.WaitGroup
	// worker by chatGPT token
	for token := range g.cfg.ChatGPT.Tokens {
		wg.Add(1)
		go g.initProcessMessagesWorker(&wg, token)
	}
	g.botClient.Run()
	wg.Wait()
}

func (g *GoTBotOpenAI) initProcessMessagesWorker(wg *sync.WaitGroup, token string) {
	defer wg.Done()
	for {
		select {
		case msg, ok := <-g.msgChan:
			if !ok {
				g.quitChan <- struct{}{}
				return
			}
			g.processMessage(msg, token)
		}
	}
}

func (g *GoTBotOpenAI) processMessage(msg *message, token string) {
	if msg == nil {
		return
	}
	var (
		respBody bytes.Buffer
		fileName string
		err      error
	)
	defer func() {
		if fileName != "" {
			err = g.botClient.ReplyFile(msg.messageID, msg.chatID, respBody.Bytes(), fileName)
		} else {
			err = g.botClient.ReplyText(msg.messageID, msg.chatID, respBody.String())
		}
		if err != nil {
			g.log.Error("Reply message error:", zap.Error(err))
		}
		respBody.Reset()
	}()
	g.checkClientSession(&respBody, msg.command, msg.chatID)
	if respBody.Len() > 0 {
		return
	}
	g.switchCommands(&respBody, msg.command, msg.chatID)
	if respBody.Len() > 0 {
		return
	}
	g.checkClientJobs(&respBody, msg.chatID)
	if respBody.Len() > 0 {
		return
	}
	fileName = g.processTextMessage(&respBody, token, msg.text, msg.chatID)
}

func (g *GoTBotOpenAI) checkClientSession(respBody *bytes.Buffer, command string, chatID int64) {
	curCmd, err := g.tClients.ClientCurrentCommand(chatID)
	switch {
	case (curCmd == "" || err != nil) && (command == "" || (command != commandStart && command != commandHelp)):
		respBody.WriteString(respBodySessionIsNotExist)
	case curCmd != "" && command == commandStart:
		respBody.WriteString(respBodySessionAlreadyExist)
	}
}

func (g *GoTBotOpenAI) checkClientJobs(respBody *bytes.Buffer, chatID int64) {
	jobs, err := g.tClients.ClientCurrentJobs(chatID)
	if err != nil {
		respBody.WriteString(respBodySessionIsNotExist)
		return
	}
	if jobs >= 1 {
		respBody.WriteString(respBodyLimitJobs)
	}
}

func (g *GoTBotOpenAI) switchCommands(respBody *bytes.Buffer, command string, chatID int64) {
	if command == "" {
		return
	}
	body, ok := respBodyByCmd[command]
	if !ok {
		respBody.WriteString(respBodyUndefinedCommand)
		return
	}
	if command == commandStop {
		g.tClients.DeleteClient(chatID)
	} else if command != commandImageCustomExample && command != commandHelp {
		g.tClients.SetClientCurrentCommand(chatID, command)
	}
	respBody.WriteString(body)
}

func (g *GoTBotOpenAI) processTextMessage(respBody *bytes.Buffer, token, text string, chatID int64) (fileName string) {
	var (
		result []byte
		err    error
	)
	command, err := g.tClients.ClientCurrentCommand(chatID)
	if err != nil {
		g.log.Error("Get client current command error:", zap.Error(err))
		respBody.WriteString(respBodySessionIsNotExist)
		return
	}
	if err = g.tClients.IncrementClientCurrentJobs(chatID); err != nil {
		g.log.Error("Increment client current jobs error:", zap.Error(err))
		respBody.WriteString(respBodySessionIsNotExist)
		return
	}
	defer func() {
		if err != nil {
			g.log.Error("OpenAI generation error:", zap.Error(err))
			if body, ok := respErrBodyByCmd[command]; ok {
				respBody.WriteString(body)
			} else if command == commandImageCustom {
				respBody.WriteString(respErrBodyCommandImageCustom(err))
			}
		} else {
			respBody.Write(result)
		}
		if err = g.tClients.DecrementClientCurrentJobs(chatID); err != nil {
			g.log.Error("Decrement client current jobs error:", zap.Error(err))
		}
	}()
	ctx := context.Background()
	switch command {
	case commandText:
		result, err = g.chatGPT.GenerateText(ctx, token, text)
	case commandImageSize256x256, commandImageSize512x512, commandImageSize1024x1024:
		result, fileName, err = g.chatGPT.GenerateImage(ctx, token, text, resolutionByImgCommand[command])
	case commandImageCustom:
		result, fileName, err = g.dreamBoothAPI.TextToImage(ctx, NewSerializedDBBodyRequest(g.cfg.DreamBooth.Key, text))
	default:
		result = []byte(respBodyUndefinedGeneration)
	}
	return
}

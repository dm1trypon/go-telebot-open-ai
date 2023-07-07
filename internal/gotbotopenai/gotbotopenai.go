package gotbotopenai

import (
	"bytes"
	"context"
	"sync"

	"go.uber.org/zap"

	"github.com/dm1trypon/go-telebot-open-ai/pkg/strgen"
)

const (
	commandStart              = "start"
	commandStop               = "stop"
	commandText               = "text"
	commandImageSize256x256   = "image256x256"
	commandImageSize512x512   = "image512x512 "
	commandImageSize1024x1024 = "image1024x1024"
	commandHelp               = "help"

	lenGenImageName    = 16
	formatGenImageName = ".jpeg"
)

type commandByChatID struct {
	value map[int64]string
	mutex sync.RWMutex
}

func (c *commandByChatID) SetCurrentCommand(chatID int64, command string) {
	defer c.mutex.Unlock()
	c.mutex.Lock()
	c.value[chatID] = command
}

func (c *commandByChatID) CurrentCommand(chatID int64) string {
	defer c.mutex.RUnlock()
	c.mutex.RLock()
	return c.value[chatID]
}

func (c *commandByChatID) DeleteCurrentCommand(chatID int64) {
	defer c.mutex.Unlock()
	c.mutex.Lock()
	delete(c.value, chatID)
}

type GoTBotOpenAI struct {
	cfg             *Config
	botClient       BotClient
	chatGPT         *ChatGPT
	commandByChatID commandByChatID
	log             *zap.Logger
	msgChan         chan *message
	quitChan        chan<- struct{}
}

func NewGoTBotOpenAI(cfg *Config, log *zap.Logger) (*GoTBotOpenAI, error) {
	msgChan := make(chan *message, cfg.LenMessageChan)
	quitChan := make(chan struct{}, 1)
	telegram, err := NewTelegram(cfg.Telegram, log, msgChan, quitChan)
	if err != nil {
		return nil, err
	}
	return &GoTBotOpenAI{
		cfg:             cfg,
		botClient:       telegram,
		chatGPT:         NewChatGPT(cfg.ChatGPT),
		commandByChatID: commandByChatID{value: make(map[int64]string)},
		log:             log,
		msgChan:         msgChan,
		quitChan:        quitChan,
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
		isFile   bool
		err      error
	)
	defer func() {
		if isFile {
			err = g.botClient.ReplyFile(msg.messageID, msg.chatID, respBody.Bytes(), strgen.Generate(lenGenImageName)+formatGenImageName)
		} else {
			err = g.botClient.ReplyText(msg.messageID, msg.chatID, respBody.String())
		}
		if err != nil {
			g.log.Error("Reply message error:", zap.Error(err))
		}
		respBody.Reset()
	}()
	g.switchCommands(&respBody, msg)
	if respBody.Len() > 0 {
		return
	}
	isFile = g.processTextMessage(&respBody, token, msg.text, msg.chatID)
}

func (g *GoTBotOpenAI) switchCommands(respBody *bytes.Buffer, msg *message) {
	if msg.command == "" {
		return
	}
	switch msg.command {
	case commandStart:
		if g.commandByChatID.CurrentCommand(msg.chatID) != "" {
			respBody.WriteString("Сессия с ботом уже активна. Чтобы посмотреть описание команд, введите команду /help.")
			return
		}
		g.commandByChatID.SetCurrentCommand(msg.chatID, msg.command)
		respBody.WriteString("Сессия с ботом активна. Чтобы посмотреть описание команд, введите команду /help.")
		return
	case commandStop:
		if g.commandByChatID.CurrentCommand(msg.chatID) == "" {
			respBody.WriteString("Сессия с ботом не активна. Чтобы начать сессию с ботом, введите команду /start. Чтобы посмотреть описание команд, введите команду /help.")
			return
		}
		g.commandByChatID.DeleteCurrentCommand(msg.chatID)
		respBody.WriteString("Сессия с ботом завершена. Возвращайтесь!")
		return
	case commandText:
		if g.commandByChatID.CurrentCommand(msg.chatID) == "" {
			respBody.WriteString("Сессия с ботом не активна. Чтобы начать сессию с ботом, введите команду /start. Чтобы посмотреть описание команд, введите команду /help.")
			return
		}
		g.commandByChatID.SetCurrentCommand(msg.chatID, msg.command)
		respBody.WriteString("Выбрана генерация текста. Введите запрос как можно подробнее, чтобы получить наиболее удовлетворительный сгенерированный текстовый ответ.")
		return
	case commandImageSize256x256:
		if g.commandByChatID.CurrentCommand(msg.chatID) == "" {
			respBody.WriteString("Сессия с ботом не активна. Чтобы начать сессию с ботом, введите команду /start. Чтобы посмотреть описание команд, введите команду /help.")
			return
		}
		g.commandByChatID.SetCurrentCommand(msg.chatID, msg.command)
		respBody.WriteString("Выбрана генерация изображений размером 256x256. Введите запрос как можно подробнее, чтобы получить наиболее удовлетворительное сгенерированное изображение.")
		return
	case commandImageSize512x512:
		if g.commandByChatID.CurrentCommand(msg.chatID) == "" {
			respBody.WriteString("Сессия с ботом не активна. Чтобы начать сессию с ботом, введите команду /start. Чтобы посмотреть описание команд, введите команду /help.")
			return
		}
		g.commandByChatID.SetCurrentCommand(msg.chatID, msg.command)
		respBody.WriteString("Выбрана генерация изображений размером 512x512. Введите запрос как можно подробнее, чтобы получить наиболее удовлетворительное сгенерированное изображение.")
		return
	case commandImageSize1024x1024:
		if g.commandByChatID.CurrentCommand(msg.chatID) == "" {
			respBody.WriteString("Сессия с ботом не активна. Чтобы начать сессию с ботом, введите команду /start. Чтобы посмотреть описание команд, введите команду /help.")
			return
		}
		g.commandByChatID.SetCurrentCommand(msg.chatID, msg.command)
		respBody.WriteString("Выбрана генерация изображений размером 1024x1024. Введите запрос как можно подробнее, чтобы получить наиболее удовлетворительное сгенерированное изображение.")
		return
	case commandHelp:
		respBody.WriteString("Доступные команды бота:\n/start - начало сессии с ботом\n/stop - завершение сессии с ботом\n/image256x256 - генерация изображений размером 256x256, используя модель OpenAI DALL·E\n/image512x512 - генерация изображений размером 512x512, используя модель OpenAI DALL·E\n/image1024x1024 - генерация изображений размером 1024x1024, используя модель OpenAI DALL·E\n/text - генерация текста, используя модель OpenAI gpt-4-32k-0613.")
	}
}

func (g *GoTBotOpenAI) processTextMessage(respBody *bytes.Buffer, token, text string, chatID int64) (isFile bool) {
	if g.commandByChatID.CurrentCommand(chatID) == "" {
		respBody.WriteString("Сессия с ботом не активна. Чтобы начать сессию с ботом, введите команду /start. Чтобы посмотреть описание команд, введите команду /help.")
		return
	}
	var (
		result []byte
		err    error
	)
	defer func() {
		if err != nil {
			g.log.Error("ChatGPT generation error:", zap.Error(err))
			respBody.WriteString("Запрос не удовлетворяет политике работы с OpenAI https://openai.com/policies/usage-policies. Пожалуйста, переформулируйте запрос.")
			isFile = false
			return
		}
		respBody.Write(result)
	}()
	ctx := context.Background()
	switch g.commandByChatID.CurrentCommand(chatID) {
	case commandText:
		result, err = g.chatGPT.GenerateText(ctx, token, text)
		return
	case commandImageSize256x256:
		result, err = g.chatGPT.GenerateImage(ctx, token, text, 1)
		isFile = true
		return
	case commandImageSize512x512:
		result, err = g.chatGPT.GenerateImage(ctx, token, text, 2)
		isFile = true
		return
	case commandImageSize1024x1024:
		result, err = g.chatGPT.GenerateImage(ctx, token, text, 3)
		isFile = true
		return
	}
	respBody.WriteString("Не выбрано, что генерировать: текст или изображение. Чтобы посмотреть описание команд, введите команду /help.")
	return
}

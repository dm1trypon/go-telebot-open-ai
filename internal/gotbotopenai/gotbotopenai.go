package gotbotopenai

import (
	"bytes"
	"context"
	"github.com/dm1trypon/go-telebot-open-ai/pkg/strgen"
	"log"
	"sync"
)

const (
	commandStart              = "start"
	commandStop               = "stop"
	commandText               = "text"
	commandImageSize256x256   = "image256x256"
	commandImageSize512x512   = "image512x512 "
	commandImageSize1024x1024 = "image1024x1024"
	commandHelp               = "help"
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
	msgChan         chan *message
	quitChan        chan<- struct{}
}

func NewGoTBotOpenAI(cfg *Config, quitChan chan<- struct{}) (*GoTBotOpenAI, error) {
	msgChan := make(chan *message)
	telegram, err := NewTelegram(cfg.Telegram, msgChan, quitChan)
	if err != nil {
		return nil, err
	}
	return &GoTBotOpenAI{
		cfg:             cfg,
		botClient:       telegram,
		chatGPT:         NewChatGPT(cfg.ChatGPT.Token),
		commandByChatID: commandByChatID{value: make(map[int64]string)},
		msgChan:         msgChan,
		quitChan:        quitChan,
	}, nil
}

func (g *GoTBotOpenAI) Run() {
	var wg sync.WaitGroup
	wg.Add(1)
	go g.initProcessMessagesWorker(&wg)
	g.botClient.Run()
	wg.Wait()
}

func (g *GoTBotOpenAI) initProcessMessagesWorker(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case msg, ok := <-g.msgChan:
			if !ok {
				g.quitChan <- struct{}{}
				return
			}
			g.processMessage(msg)
		}
	}
}

func (g *GoTBotOpenAI) processMessage(msg *message) {
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
			err = g.botClient.ReplyFile(msg.messageID, msg.chatID, respBody.Bytes(), strgen.Generate()+".jpeg")
		} else {
			err = g.botClient.ReplyText(msg.messageID, msg.chatID, respBody.String())
		}
		if err != nil {
			log.Println("Reply message err:", err)
		}
		respBody.Reset()
	}()
	g.switchCommands(&respBody, msg)
	if respBody.Len() > 0 {
		return
	}
	isFile = g.processTextMessage(&respBody, msg.text, msg.chatID)
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

func (g *GoTBotOpenAI) processTextMessage(respBody *bytes.Buffer, text string, chatID int64) (isFile bool) {
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
			log.Println("ChatGPT generation error:", err)
			respBody.WriteString("Запрос не удовлетворяет политике работы с OpenAI https://openai.com/policies/usage-policies. Пожалуйста, переформулируйте запрос.")
			isFile = false
			return
		}
		respBody.Write(result)
	}()
	ctx := context.Background()
	switch g.commandByChatID.CurrentCommand(chatID) {
	case commandText:
		result, err = g.chatGPT.GenerateText(ctx, text)
		return
	case commandImageSize256x256:
		result, err = g.chatGPT.GenerateImage(ctx, text, 1)
		isFile = true
		return
	case commandImageSize512x512:
		result, err = g.chatGPT.GenerateImage(ctx, text, 2)
		isFile = true
		return
	case commandImageSize1024x1024:
		result, err = g.chatGPT.GenerateImage(ctx, text, 3)
		isFile = true
		return
	}
	respBody.WriteString("Не выбрано, что генерировать: текст или изображение. Чтобы посмотреть описание команд, введите команду /help.")
	return
}

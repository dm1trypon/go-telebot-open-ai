package tbotopenai

import (
	"sync"

	"go.uber.org/zap"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// NewUpdate gets updates since the last Offset
const updaterOffset = 0

type message struct {
	chatID    int64
	messageID int
	text      string
	command   string
	username  string
}

type Messenger interface {
	Run()
	Stop()
	ReplyText(int, int64, string) error
	ReplyFile(int, int64, []byte, string) error
}

type Telegram struct {
	bot          *tgbotapi.BotAPI
	updateConfig tgbotapi.UpdateConfig
	updateChan   tgbotapi.UpdatesChannel
	log          *zap.Logger
	msgChan      chan<- *message
}

func NewTelegram(cfg *TelegramSettings, log *zap.Logger, msgChan chan<- *message) (*Telegram, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return nil, err
	}
	bot.Debug = cfg.Debug
	updateConfig := tgbotapi.NewUpdate(updaterOffset)
	updateConfig.Timeout = cfg.Timeout
	updateChan := bot.GetUpdatesChan(updateConfig)
	return &Telegram{bot, updateConfig, updateChan, log, msgChan}, nil
}

func (t *Telegram) Run() {
	var wg sync.WaitGroup
	wg.Add(1)
	go t.initReadingMessagesWorker(&wg)
	wg.Wait()
}

func (t *Telegram) Stop() {
	t.bot.StopReceivingUpdates()
}

func (t *Telegram) initReadingMessagesWorker(wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			t.log.Error("Recovered panic err:", zap.Any("panic", r))
		}
	}()
	defer wg.Done()
	for {
		select {
		case update, ok := <-t.updateChan:
			if !ok {
				return
			}
			if update.Message == nil || update.Message.Chat == nil {
				continue
			}
			t.msgChan <- &message{
				chatID:    update.Message.Chat.ID,
				messageID: update.Message.MessageID,
				text:      update.Message.Text,
				command:   update.Message.Command(),
				username:  update.Message.From.UserName,
			}
		}
	}
}

func (t *Telegram) ReplyText(messageID int, chatID int64, body string) (err error) {
	msg := tgbotapi.NewMessage(chatID, body)
	msg.ReplyToMessageID = messageID
	_, err = t.bot.Send(msg)
	return
}

func (t *Telegram) ReplyFile(messageID int, chatID int64, body []byte, fileName string) (err error) {
	fb := tgbotapi.FileBytes{
		Name:  fileName,
		Bytes: body,
	}
	docCfg := tgbotapi.NewDocument(chatID, fb)
	docCfg.ReplyToMessageID = messageID
	_, err = t.bot.Send(docCfg)
	return
}

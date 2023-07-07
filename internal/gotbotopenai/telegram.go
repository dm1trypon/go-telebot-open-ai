package gotbotopenai

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

// NewUpdate gets updates since the last Offset
const updaterOffset = 0

type message struct {
	chatID    int64
	messageID int
	text      string
	command   string
}

type BotClient interface {
	Run()
	ReplyText(int, int64, string) error
	ReplyFile(int, int64, []byte, string) error
}

type Telegram struct {
	bot          *tgbotapi.BotAPI
	updateConfig tgbotapi.UpdateConfig
	updateChan   tgbotapi.UpdatesChannel
	log          *zap.Logger
	msgChan      chan *message
	quitChan     chan<- struct{}
}

func NewTelegram(cfg *TelegramSettings, log *zap.Logger, msgChan chan *message, quitChan chan<- struct{}) (*Telegram, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return nil, err
	}
	bot.Debug = cfg.Debug
	updateConfig := tgbotapi.NewUpdate(updaterOffset)
	updateConfig.Timeout = cfg.Timeout
	updateChan := bot.GetUpdatesChan(updateConfig)
	return &Telegram{bot, updateConfig, updateChan, log, msgChan, quitChan}, nil
}

func (t *Telegram) Run() {
	go t.initReadingMessagesWorker()
}

func (t *Telegram) initReadingMessagesWorker() {
	for {
		select {
		case update, ok := <-t.updateChan:
			if !ok {
				// shutdown service
				t.quitChan <- struct{}{}
				return
			}
			if update.Message == nil || update.Message.Chat == nil {
				continue
			}
			t.log.Debug("Received message",
				zap.String("user", update.Message.From.UserName),
				zap.String("body", update.Message.Text),
				zap.String("command", update.Message.Command()))
			if len(t.msgChan) == cap(t.msgChan) {
				if err := t.ReplyText(update.Message.MessageID, update.Message.Chat.ID, "Слишком большая нагрузка на бота, выполните запрос чуть позже"); err != nil {
					t.log.Error("Reply message error:", zap.Error(err))
				}
				return
			}
			t.msgChan <- &message{update.Message.Chat.ID, update.Message.MessageID, update.Message.Text, update.Message.Command()}
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

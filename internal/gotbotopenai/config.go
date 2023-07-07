package gotbotopenai

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var errEmptyChatGPTTokens = errors.New("empty ChatGPT tokens")

type Config struct {
	Telegram       TelegramSettings
	ChatGPT        ChatGPTSettings
	Logger         zap.Config
	LenMessageChan int
}

type TelegramSettings struct {
	Token             string
	Debug             bool
	Timeout           int
	ReconnectInterval int // seconds
}

type ChatGPTSettings struct {
	Tokens       map[string]struct{}
	RetryRequest int
	RetryTimeout int // seconds
}

func NewConfig() (*Config, error) {
	telegramToken, err := strconv.ParseBool(os.Getenv("TELEGRAM_DEBUG"))
	if err != nil {
		return nil, err
	}

	telegramTimeout, err := strconv.Atoi(os.Getenv("TELEGRAM_TIMEOUT"))
	if err != nil {
		return nil, err
	}
	telegramReconnectInterval, err := strconv.Atoi(os.Getenv("TELEGRAM_RECONNECT_INTERVAL"))
	if err != nil {
		return nil, err
	}
	logOutputPath := os.Getenv("LOG_OUTPUT_PATH")
	if logOutputPath != "stdout" {
		if err = os.MkdirAll(filepath.Dir(logOutputPath), os.ModePerm); err != nil {
			return nil, err
		}
	}
	lenMessageChan, err := strconv.Atoi(os.Getenv("LEN_MESSAGE_CHAN"))
	if err != nil {
		return nil, err
	}
	chatGPTTokens := os.Getenv("CHAT_GPT_TOKENS")
	chatGPTTokens = strings.TrimSuffix(chatGPTTokens, "\n")
	chatGPTTokens = strings.TrimSuffix(chatGPTTokens, "\r")
	tokensArr := strings.Split(os.Getenv("CHAT_GPT_TOKENS"), ",")
	// to exclude identical tokens
	tokensMap := make(map[string]struct{}, len(tokensArr))
	for idx := range tokensArr {
		if tokensArr[idx] == "" {
			continue
		}
		tokensMap[tokensArr[idx]] = struct{}{}
	}
	if len(tokensMap) == 0 {
		return nil, errEmptyChatGPTTokens
	}
	chatGPTRetryRequest, err := strconv.Atoi(os.Getenv("CHAT_GPT_RETRY_REQUEST"))
	if err != nil {
		return nil, err
	}
	chatGPTRetryTimeout, err := strconv.Atoi(os.Getenv("CHAT_GPT_RETRY_TIMEOUT"))
	if err != nil {
		return nil, err
	}
	return &Config{
		Telegram: TelegramSettings{
			Token:             os.Getenv("TELEGRAM_TOKEN"),
			Debug:             telegramToken,
			Timeout:           telegramTimeout,
			ReconnectInterval: telegramReconnectInterval,
		},
		ChatGPT: ChatGPTSettings{
			Tokens:       tokensMap,
			RetryRequest: chatGPTRetryRequest,
			RetryTimeout: chatGPTRetryTimeout,
		},
		Logger:         newLogger(),
		LenMessageChan: lenMessageChan,
	}, nil
}

func newLogger() zap.Config {
	zapCfg := zap.NewDevelopmentConfig()
	zapCfg.Level.SetLevel(getLogLevel())
	zapCfg.Encoding = "json"
	zapCfg.OutputPaths = []string{os.Getenv("LOG_OUTPUT_PATH")}
	zapCfg.EncoderConfig = zap.NewDevelopmentEncoderConfig()
	zapCfg.EncoderConfig.MessageKey = "msg"
	zapCfg.EncoderConfig.LevelKey = "level"
	zapCfg.EncoderConfig.TimeKey = "dttm"
	zapCfg.EncoderConfig.CallerKey = "call"
	zapCfg.EncoderConfig.StacktraceKey = "stack_trace_key"
	zapCfg.EncoderConfig.NameKey = "name_key"
	zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	return zapCfg
}

func getLogLevel() zapcore.Level {
	levelStr := os.Getenv("LOG_LEVEL")
	var level zapcore.Level
	switch levelStr {
	case "DEBUG":
		level = zap.DebugLevel
	case "INFO":
		level = zap.InfoLevel
	case "ERROR":
		level = zap.ErrorLevel
	default:
		level = zap.InfoLevel
	}
	return level
}

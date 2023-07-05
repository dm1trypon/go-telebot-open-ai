package gotbotopenai

import (
	"os"
	"path/filepath"
	"strconv"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Telegram TelegramSettings
	ChatGPT  ChatGPTSettings
	Logger   zap.Config
}

type TelegramSettings struct {
	Token             string
	Debug             bool
	Timeout           int
	ReconnectInterval int // seconds
}

type ChatGPTSettings struct {
	Token string
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
	return &Config{
		Telegram: TelegramSettings{
			Token:             os.Getenv("TELEGRAM_TOKEN"),
			Debug:             telegramToken,
			Timeout:           telegramTimeout,
			ReconnectInterval: telegramReconnectInterval,
		},
		ChatGPT: ChatGPTSettings{
			Token: os.Getenv("CHAT_GPT_TOKEN"),
		},
		Logger: newLogger(),
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

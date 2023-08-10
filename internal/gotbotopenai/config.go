package gotbotopenai

import (
	"errors"
	"github.com/dm1trypon/go-telebot-open-ai/pkg/gdapi"
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
	DreamBooth     DreamBoothSettings
	GoogleDriveAPI GoogleDriveAPI
	Logger         zap.Config
	LenMessageChan int
	MaxClientsJobs int
}

type TelegramSettings struct {
	Token             string
	Debug             bool
	Timeout           int
	ReconnectInterval int // seconds
}

type ChatGPTSettings struct {
	Tokens        map[string]struct{}
	RetryRequest  int
	RetryInterval int // seconds
}

type DreamBoothSettings struct {
	ConfigFileName string
	RetryCount     int
	RetryInterval  int
}

type GoogleDriveAPI struct {
	CredentialsSettings gdapi.CredentialsSettings
	TokenSettings       gdapi.TokenSettings
	Timeout             int
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
	chatGPTRetryInterval, err := strconv.Atoi(os.Getenv("CHAT_GPT_RETRY_INTERVAL"))
	if err != nil {
		return nil, err
	}
	dreamBoothRetryCount, err := strconv.Atoi(os.Getenv("DREAMBOOTH_RETRY_COUNT"))
	if err != nil {
		return nil, err
	}
	dreamBoothRetryInterval, err := strconv.Atoi(os.Getenv("DREAMBOOTH_RETRY_INTERVAL"))
	if err != nil {
		return nil, err
	}
	maxClientsJobs, err := strconv.Atoi(os.Getenv("MAX_CLIENTS_JOBS"))
	if err != nil {
		return nil, err
	}
	gdAPITimeout, err := strconv.Atoi(os.Getenv("GD_API_TIMEOUT"))
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
			Tokens:        tokensMap,
			RetryRequest:  chatGPTRetryRequest,
			RetryInterval: chatGPTRetryInterval,
		},
		DreamBooth: DreamBoothSettings{
			ConfigFileName: os.Getenv("DREAMBOOTH_CONFIG_FILENAME"),
			RetryCount:     dreamBoothRetryCount,
			RetryInterval:  dreamBoothRetryInterval,
		},
		GoogleDriveAPI: GoogleDriveAPI{
			CredentialsSettings: gdapi.CredentialsSettings{
				ClientID:                os.Getenv("GD_CLIENT_ID"),
				ProjectID:               os.Getenv("GD_PROJECT_ID"),
				AuthURI:                 os.Getenv("GD_AUTH_URI"),
				TokenURI:                os.Getenv("GD_TOKEN_URI"),
				AuthProviderX509CertURL: os.Getenv("GD_AUTH_PROVIDER_X509_CERT_URL"),
				ClientSecret:            os.Getenv("GD_CLIENT_SECRET"),
			},
			TokenSettings: gdapi.TokenSettings{
				AccessToken:  os.Getenv("GD_ACCESS_TOKEN"),
				TokenType:    os.Getenv("GD_TOKEN_TYPE"),
				RefreshToken: os.Getenv("GD_REFRESH_TOKEN"),
				Expiry:       os.Getenv("GD_EXPIRY"),
			},
			Timeout: gdAPITimeout,
		},
		Logger:         newLogger(logOutputPath),
		LenMessageChan: lenMessageChan,
		MaxClientsJobs: maxClientsJobs,
	}, nil
}

func newLogger(logOutputPath string) zap.Config {
	zapCfg := zap.NewDevelopmentConfig()
	zapCfg.Level.SetLevel(getLogLevel())
	zapCfg.Encoding = "json"
	zapCfg.OutputPaths = []string{logOutputPath}
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

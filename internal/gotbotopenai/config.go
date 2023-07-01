package gotbotopenai

type Config struct {
	Telegram TelegramSettings
	ChatGPT  ChatGPTSettings
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

func NewConfig() *Config {
	return &Config{
		Telegram: TelegramSettings{
			Token:             "6339322764:AAGXPnK3BDqYKRuvXP6JUghl4ffh5xkaV4A",
			Debug:             true,
			Timeout:           100,
			ReconnectInterval: 1,
		},
		ChatGPT: ChatGPTSettings{
			Token: "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",
		},
	}
}

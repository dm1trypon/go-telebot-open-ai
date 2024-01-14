package tbotopenai

import (
	"io/ioutil"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Telegram                TelegramSettings   `yaml:"telegram"`
	ChatGPT                 ChatGPTSettings    `yaml:"chatgpt"`
	OpenAI                  OpenAISettings     `yaml:"openai"`
	DreamBooth              DreamBoothSettings `yaml:"dreambooth"`
	Logger                  zap.Config         `yaml:"log"`
	LenMessageChan          int                `yaml:"len_message_chan"`
	LenQueueTaskChan        int                `yaml:"len_queue_task_chan"`
	QueueMessageWorkers     int                `yaml:"queue_message_workers"`
	MaxClientOpenAIJobs     int                `yaml:"max_client_openai_jobs"`
	MaxClientChatGPTJobs    int                `yaml:"max_client_chatgpt_jobs"`
	MaxClientDreamBoothJobs int                `yaml:"max_client_dreambooth_jobs"`
}

type TelegramSettings struct {
	Token             string `yaml:"token"`
	Debug             bool   `yaml:"debug"`
	Timeout           int    `yaml:"timeout"`
	ReconnectInterval int    `yaml:"reconnect_interval"`
}

type OpenAISettings struct {
	Token         string `yaml:"token"`
	RetryCount    int    `yaml:"retry_count"`
	RetryInterval int    `yaml:"retry_interval"`
	Timeout       int    `yaml:"timeout"`
}

type ChatGPTSettings struct {
	Timeout int `yaml:"timeout"`
}

type DreamBoothSettings struct {
	Tokens        []string `yaml:"tokens"`
	RetryInterval int      `yaml:"retry_interval"`
	Timeout       int      `yaml:"timeout"`
}

func NewConfig(filePath string) (*Config, error) {
	cfg, err := readConfig(filePath)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func readConfig(filename string) (*Config, error) {
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = yaml.Unmarshal(yamlFile, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

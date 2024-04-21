package tbotopenai

import (
	"io/ioutil"
	"time"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Telegram                TelegramSettings    `yaml:"telegram"`
	ChatGPT                 ChatGPTSettings     `yaml:"chatgpt"`
	OpenAI                  OpenAISettings      `yaml:"openai"`
	DreamBooth              DreamBoothSettings  `yaml:"dreambooth"`
	FusionBrain             FusionBrainSettings `yaml:"fusionbrain"`
	Roles                   RolesSettings       `yaml:"roles"`
	Permissions             PermissionSettings  `yaml:"permissions"`
	Stats                   StatsSettings       `yaml:"stats"`
	Logger                  zap.Config          `yaml:"log"`
	LenMessageChan          int                 `yaml:"len_message_chan"`
	LenQueueTaskChan        int                 `yaml:"len_queue_task_chan"`
	QueueMessageWorkers     int                 `yaml:"queue_message_workers"`
	MaxClientOpenAIJobs     int                 `yaml:"max_client_openai_jobs"`
	MaxClientChatGPTJobs    int                 `yaml:"max_client_chatgpt_jobs"`
	MaxClientDreamBoothJobs int                 `yaml:"max_client_dreambooth_jobs"`
	MaxLogRows              int                 `yaml:"max_log_rows"`
	PathBlackList           string              `yaml:"path_blacklist"`
}

type TelegramSettings struct {
	Token   string `yaml:"token"`
	Debug   bool   `yaml:"debug"`
	Timeout int    `yaml:"timeout"`
}

type OpenAISettings struct {
	Token         string        `yaml:"token"`
	RetryCount    int           `yaml:"retry_count"`
	RetryInterval time.Duration `yaml:"retry_interval"`
	Timeout       time.Duration `yaml:"timeout"`
}

type ChatGPTSettings struct {
	Timeout time.Duration `yaml:"timeout"`
}

type DreamBoothSettings struct {
	Tokens        []string      `yaml:"tokens"`
	RetryInterval time.Duration `yaml:"retry_interval"`
	Timeout       time.Duration `yaml:"timeout"`
}

type FusionBrainSettings struct {
	RetryInterval time.Duration `yaml:"retry_interval"`
	Timeout       time.Duration `yaml:"timeout"`
	Key           string        `yaml:"key"`
	SecretKey     string        `yaml:"secret_key"`
}

type RolesSettings struct {
	Admins []string `yaml:"admin"`
	Users  []string `yaml:"user"`
}

type PermissionSettings struct {
	AdminCommands []string `yaml:"admin"`
	UserCommands  []string `yaml:"user"`
}

type StatsSettings struct {
	Interval time.Duration `yaml:"interval"`
	Filepath string        `yaml:"filepath"`
}

func NewConfig(filename string) (*Config, error) {
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

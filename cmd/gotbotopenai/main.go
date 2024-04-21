package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/dm1trypon/go-telebot-open-ai/internal/tbotopenai"
)

func main() {
	configPath := flag.String("c", "config.yaml", "Путь до файла конфигурации")
	flag.Parse()
	cfg, err := tbotopenai.NewConfig(*configPath)
	if err != nil {
		log.Fatalf("Reading config error: %v", err)
	}
	logger, err := cfg.Logger.Build()
	if err != nil {
		log.Fatalf("Can not create logger: %v", err)
	}
	defer func() {
		if err = logger.Sync(); err != nil {
			log.Fatalf("logger sync err: %v", err)
		}
	}()
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Recovered panic err:", zap.Any("panic", r))
		}
	}()
	tBotOpenAI, err := tbotopenai.NewTBotOpenAI(cfg, logger)
	if err != nil {
		log.Fatal(err)
	}
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		logger.Info("Caught signal, terminating", zap.String("signal", sig.String()))
		tBotOpenAI.Stop()
	}()
	logger.Info("Starting OpenAI bot")
	tBotOpenAI.Run()
	logger.Info("Stopping OpenAI bot")
}

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/dm1trypon/go-telebot-open-ai/internal/gotbotopenai"
)

func main() {
	cfg, err := gotbotopenai.NewConfig()
	if err != nil {
		log.Fatalf("Reading config error: %v", err)
	}
	logger, err := cfg.Logger.Build()
	if err != nil {
		log.Fatalf("Can not create logger: %v", err)
	}
	defer logger.Sync()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		logger.Info("Caught signal, terminating", zap.String("signal", sig.String()))
	}()
	goTBotOpenAi, err := gotbotopenai.NewGoTBotOpenAI(cfg, logger)
	if err != nil {
		log.Fatal(err)
	}
	logger.Info("Starting OpenAI bot")
	goTBotOpenAi.Run()
	logger.Info("Stopping OpenAI bot")
}

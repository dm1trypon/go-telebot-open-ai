package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/dm1trypon/go-telebot-open-ai/internal/gotbotopenai"
)

// 6339322764:AAGXPnK3BDqYKRuvXP6JUghl4ffh5xkaV4A
// 5930839504:AAEKsOufyhnQuwL3kOTJJQyHzJnJqAWY0GU
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

	goTBotOpenAi, err := gotbotopenai.NewGoTBotOpenAI(cfg, logger)
	if err != nil {
		log.Fatal(err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		logger.Info("Caught signal, terminating", zap.String("signal", sig.String()))
	}()

	logger.Info("Starting OpenAI bot")
	goTBotOpenAi.Run()
	logger.Info("Stopping OpenAI bot")
}

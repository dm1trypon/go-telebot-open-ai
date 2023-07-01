package main

import (
	"github.com/dm1trypon/go-telebot-open-ai/internal/gotbotopenai"
	"log"
)

func main() {
	quitChan := make(chan struct{}, 1)
	goTBotOpenAi, err := gotbotopenai.NewGoTBotOpenAI(gotbotopenai.NewConfig(), quitChan)
	if err != nil {
		log.Fatal(err)
	}
	goTBotOpenAi.Run()
}

package tbotopenai

import (
	"context"

	"github.com/dm1trypon/go-telebot-open-ai/pkg/chatgptfree"
)

type ChatGPTBot struct{}

func NewChatGPTBot() *ChatGPTBot {
	return &ChatGPTBot{}
}

func (c *ChatGPTBot) GenerateText(ctx context.Context, prompt string) ([]byte, error) {
	return chatgptfree.GenerateText(ctx, prompt)
}

func (c *ChatGPTBot) GenerateImage(_ context.Context, _ string) ([]byte, string, error) {
	return nil, "", nil
}

package gotbotopenai

import (
	"context"
	"encoding/base64"
	"errors"

	"github.com/sashabaranov/go-openai"
)

var imageSizes = map[int]string{
	1: openai.CreateImageSize256x256,
	2: openai.CreateImageSize512x512,
	3: openai.CreateImageSize1024x1024,
}

type ChatGPT struct {
	clients map[string]*openai.Client
}

func NewChatGPT(tokens map[string]struct{}) *ChatGPT {
	chatGPT := &ChatGPT{
		clients: make(map[string]*openai.Client, len(tokens)),
	}
	for token := range tokens {
		chatGPT.clients[token] = openai.NewClient(token)
	}
	return chatGPT
}

func (c *ChatGPT) GenerateImage(ctx context.Context, token, prompt string, sizeType int) ([]byte, error) {
	reqBase64 := openai.ImageRequest{
		Prompt:         prompt,
		Size:           imageSizes[sizeType],
		ResponseFormat: openai.CreateImageResponseFormatB64JSON,
		N:              1,
	}
	respBase64, err := c.clients[token].CreateImage(ctx, reqBase64)
	if err != nil {
		return nil, err
	}
	if len(respBase64.Data) == 0 {
		return nil, errors.New("empty resp data")
	}
	return base64.StdEncoding.DecodeString(respBase64.Data[0].B64JSON)
}

func (c *ChatGPT) GenerateText(ctx context.Context, token, content string) ([]byte, error) {
	resp, err := c.clients[token].CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo0613,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: content,
				},
			},
		},
	)
	if err != nil {
		return nil, err
	}
	if len(resp.Choices) == 0 {
		return nil, errors.New("empty resp choices")
	}
	return []byte(resp.Choices[0].Message.Content), nil
}

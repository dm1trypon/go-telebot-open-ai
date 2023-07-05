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
	client *openai.Client
}

func NewChatGPT(token string) *ChatGPT {
	return &ChatGPT{
		client: openai.NewClient(token),
	}
}

func (c *ChatGPT) GenerateImage(ctx context.Context, prompt string, sizeType int) ([]byte, error) {
	reqBase64 := openai.ImageRequest{
		Prompt:         prompt,
		Size:           imageSizes[sizeType],
		ResponseFormat: openai.CreateImageResponseFormatB64JSON,
		N:              1,
	}
	respBase64, err := c.client.CreateImage(ctx, reqBase64)
	if err != nil {
		return nil, err
	}
	if len(respBase64.Data) == 0 {
		return nil, errors.New("empty resp data")
	}
	return base64.StdEncoding.DecodeString(respBase64.Data[0].B64JSON)
}

func (c *ChatGPT) GenerateText(ctx context.Context, content string) ([]byte, error) {
	resp, err := c.client.CreateChatCompletion(
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

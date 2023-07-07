package gotbotopenai

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
)

const (
	errChatGPTStatusCode503Response = "status code: 503"
	errChatGPTStatusCode429Response = "status code: 429"
)

var imageSizes = map[int]string{
	1: openai.CreateImageSize256x256,
	2: openai.CreateImageSize512x512,
	3: openai.CreateImageSize1024x1024,
}

type ChatGPT struct {
	clients      map[string]*openai.Client
	retryRequest int
	retryTimeout int
}

func NewChatGPT(cfg ChatGPTSettings) *ChatGPT {
	chatGPT := &ChatGPT{
		clients:      make(map[string]*openai.Client, len(cfg.Tokens)),
		retryRequest: cfg.RetryRequest,
		retryTimeout: cfg.RetryTimeout,
	}
	for token := range cfg.Tokens {
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
	var (
		respBase64 openai.ImageResponse
		err        error
	)
	for i := 0; i < c.retryRequest; i++ {
		respBase64, err = c.clients[token].CreateImage(ctx, reqBase64)
		if isSkipRetry(err) {
			break
		}
		time.Sleep(time.Second * time.Duration(c.retryTimeout))
	}
	respBase64, err = c.clients[token].CreateImage(ctx, reqBase64)
	if err != nil {
		return nil, err
	}
	if len(respBase64.Data) == 0 {
		return nil, errors.New("empty resp data")
	}
	return base64.StdEncoding.DecodeString(respBase64.Data[0].B64JSON)
}

func (c *ChatGPT) GenerateText(ctx context.Context, token, content string) ([]byte, error) {
	var (
		resp openai.ChatCompletionResponse
		err  error
	)
	for i := 0; i < c.retryRequest; i++ {
		resp, err = c.clients[token].CreateChatCompletion(
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
		if isSkipRetry(err) {
			break
		}
		time.Sleep(time.Second * time.Duration(c.retryTimeout))
	}
	if err != nil {
		return nil, err
	}
	if len(resp.Choices) == 0 {
		return nil, errors.New("empty resp choices")
	}
	return []byte(resp.Choices[0].Message.Content), nil
}

func isSkipRetry(err error) bool {
	return err == nil || (err != nil && (!strings.Contains(err.Error(), errChatGPTStatusCode503Response) && (!strings.Contains(err.Error(), errChatGPTStatusCode429Response))))
}

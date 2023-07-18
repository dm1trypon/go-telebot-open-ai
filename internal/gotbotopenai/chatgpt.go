package gotbotopenai

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/dm1trypon/go-telebot-open-ai/pkg/strgen"
	"github.com/sashabaranov/go-openai"
)

const (
	errChatGPTStatusCode503Response = "status code: 503"
	errChatGPTStatusCode429Response = "status code: 429"

	lenImgFileName = 20
	formatImgFile  = ".png"
)

var (
	errChatGPTEmptyFileName    = errors.New("ChatGPT filename is empty")
	errChatGPTEmptyRespData    = errors.New("ChatGPT empty response data")
	errChatGPTEmptyRespChoices = errors.New("ChatGPT empty resp choices")
)

type ChatGPT struct {
	clients       map[string]*openai.Client
	retryRequest  int
	retryInterval int
}

func NewChatGPT(cfg *ChatGPTSettings) *ChatGPT {
	chatGPT := &ChatGPT{
		clients:       make(map[string]*openai.Client, len(cfg.Tokens)),
		retryRequest:  cfg.RetryRequest,
		retryInterval: cfg.RetryInterval,
	}
	for token := range cfg.Tokens {
		chatGPT.clients[token] = openai.NewClient(token)
	}
	return chatGPT
}

func (c *ChatGPT) GenerateImage(ctx context.Context, token, prompt, size string) ([]byte, string, error) {
	reqBase64 := openai.ImageRequest{
		Prompt:         prompt,
		Size:           size,
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
		time.Sleep(time.Second * time.Duration(c.retryInterval))
	}
	if len(respBase64.Data) == 0 {
		return nil, "", errChatGPTEmptyRespData
	}
	body, err := base64.StdEncoding.DecodeString(respBase64.Data[0].B64JSON)
	if err != nil {
		return nil, "", err
	}
	return body, strgen.Generate(lenImgFileName) + formatImgFile, err
}

func (c *ChatGPT) GenerateText(ctx context.Context, token, content string) ([]byte, error) {
	var (
		resp openai.ChatCompletionResponse
		err  error
	)
	req := openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo0613,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: content,
			},
		},
	}
	for i := 0; i < c.retryRequest; i++ {
		resp, err = c.clients[token].CreateChatCompletion(ctx, req)
		if isSkipRetry(err) {
			break
		}
		time.Sleep(time.Second * time.Duration(c.retryInterval))
	}
	if err != nil {
		return nil, err
	}
	if len(resp.Choices) == 0 {
		return nil, errChatGPTEmptyRespChoices
	}
	return []byte(resp.Choices[0].Message.Content), nil
}

func isSkipRetry(err error) bool {
	return err == nil || (err != nil && (!strings.Contains(err.Error(), errChatGPTStatusCode503Response) && (!strings.Contains(err.Error(), errChatGPTStatusCode429Response))))
}

package tbotopenai

// TODO: ChatGPT не используется, так как нет ключей OpenAI.

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"

	"github.com/dm1trypon/go-telebot-open-ai/pkg/strgen"
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

type OpenAI struct {
	client        *openai.Client
	retryCount    int
	retryInterval time.Duration
}

func NewOpenAI(cfg *OpenAISettings) *OpenAI {
	chatGPT := &OpenAI{
		client:        openai.NewClient(cfg.Token),
		retryCount:    cfg.RetryCount,
		retryInterval: cfg.RetryInterval,
	}
	return chatGPT
}

func (o *OpenAI) GenerateImage(ctx context.Context, prompt string) ([]byte, string, error) {
	reqBase64 := openai.ImageRequest{
		Prompt:         prompt,
		Size:           openai.CreateImageSize1024x1024,
		ResponseFormat: openai.CreateImageResponseFormatB64JSON,
		N:              1,
	}
	var (
		respBase64 openai.ImageResponse
		err        error
	)
	for i := 0; i < o.retryCount; i++ {
		respBase64, err = o.client.CreateImage(ctx, reqBase64)
		if isSkipRetry(err) {
			break
		}
		time.Sleep(o.retryInterval)
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

func (o *OpenAI) GenerateText(ctx context.Context, prompt string) ([]byte, error) {
	var (
		resp openai.ChatCompletionResponse
		err  error
	)
	req := openai.ChatCompletionRequest{
		Model: openai.GPT432K0613,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	}
	for i := 0; i < o.retryCount; i++ {
		resp, err = o.client.CreateChatCompletion(ctx, req)
		if isSkipRetry(err) {
			break
		}
		time.Sleep(o.retryInterval)
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

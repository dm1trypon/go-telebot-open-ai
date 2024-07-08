package chatgptfree

import (
	"bytes"
	"context"
	"errors"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
)

const chatGPTTextURI = "https://gpt-chatbotru-chat-main.ru/api/openai/v1/chat/completions"

var (
	errResponseCodeIsNot200 = errors.New("response code is not 200")
	errEmptyRespChoices     = errors.New("response's choices are empty")
)

func GenerateText(ctx context.Context, prompt string) ([]byte, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.SetRequestURI(chatGPTTextURI)
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://gpt-chatbotru-chat-main.ru")
	req.Header.Set("Priority", "u=1, i")
	req.Header.Set("Referer", "https://gpt-chatbotru-chat-main.ru/")
	req.Header.Set("Sec-CH-UA", `"Not/A)Brand";v="8", "Chromium";v="126", "Google Chrome";v="126"`)
	req.Header.Set("Sec-CH-UA-Mobile", "?0")
	req.Header.Set("Sec-CH-UA-Platform", `"Windows"`)
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")
	req.SetBody(prepareRequestBody(prompt))
	bodyChan := make(chan []byte, 1)
	errChan := make(chan error, 1)
	go func() {
		if err := fasthttp.Do(req, resp); err != nil {
			errChan <- err
			return
		}
		if resp.StatusCode() != fasthttp.StatusOK {
			errChan <- errResponseCodeIsNot200
			return
		}
		var p fastjson.Parser
		v, err := p.ParseBytes(resp.Body())
		if err != nil {
			errChan <- err
			return
		}
		choices := v.GetArray("choices")
		if len(choices) == 0 {
			errChan <- errEmptyRespChoices
			return
		}
		bodyChan <- choices[0].Get("message").GetStringBytes("content")
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case body := <-bodyChan:
		return body, nil
	case err := <-errChan:
		return nil, err
	}
}

func prepareRequestBody(content string) []byte {
	var b bytes.Buffer
	b.WriteString(`{"messages":[{"role":"user","content":"`)
	b.WriteString(content)
	b.WriteString(`"}],"stream":false,"model":"gpt-3.5","temperature":0.5,"presence_penalty":0,"frequency_penalty":0,"top_p":1}`)
	return b.Bytes()
}

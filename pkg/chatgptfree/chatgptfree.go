package chatgptfree

import (
	"context"
	"errors"

	"github.com/valyala/fasthttp"
)

const chatGPTTextURL = "https://chatgptbot.ru/gptbot.php"

var errResponseCodeIsNot200 = errors.New("response code is not 200")

func GenerateText(ctx context.Context, prompt string) ([]byte, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	req.SetRequestURI(chatGPTTextURL)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("authority", "chatgptbot.ru")
	req.Header.Set("accept", "*/*")
	req.Header.Set("accept-language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("origin", "https://chatgptbot.ru")
	req.Header.Set("referer", "https://chatgptbot.ru/chat/")
	req.Header.Set("sec-ch-ua", `"Chromium";v="116", "Not)A;Brand";v="24", "Google Chrome";v="116"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36")
	req.Header.Set("x-requested-with", "XMLHttpRequest")
	req.PostArgs().Set("engine", "gpt-3.5-turbo")
	req.PostArgs().Set("prompt", prompt)
	req.PostArgs().Set("temperature", "0.6")
	req.PostArgs().Set("max_tokens", "3159")
	req.PostArgs().Set("top_p", "1")
	req.PostArgs().Set("frequency_penalty", "0")
	req.PostArgs().Set("presence_penalty", "0.6")
	req.PostArgs().Set("n", "1")
	req.PostArgs().Set("stop[]", "Human:")
	req.PostArgs().Set("messages[0][role]", "user")
	req.PostArgs().Set("messages[0][content]", prompt)
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
		bodyChan <- resp.Body()
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

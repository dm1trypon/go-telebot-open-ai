package tbotopenai

import (
	"bytes"
	"context"
	"errors"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
	"go.uber.org/zap"
)

var (
	errDBStatusError                 = errors.New("DreamBooth response status error")
	errDBMonthLimit                  = errors.New("DreamBooth response month limit error")
	errDBInvalidRespCode             = errors.New("DreamBooth response status code is not 200")
	errDBParsingRespBody             = errors.New("DreamBooth parsing response body error")
	errDBRequestIDIsEmpty            = errors.New("DreamBooth 'request_id' in response body is empty")
	errDBOutputIsEmpty               = errors.New("DreamBooth 'output' in response is empty")
	errDBFQIInvalidRespCode          = errors.New("DreamBooth FetchQueuedImages response status code is not 200")
	errDBFQIParsingRespBody          = errors.New("DreamBooth FetchQueuedImages parsing response body error")
	errDBDownloadFileInvalidRespCode = errors.New("DreamBooth download file response status code is not 200")
	errDBDownloadFileRespBodyIsEmpty = errors.New("DreamBooth download file empty response body")
	errDBEmptyTokens                 = errors.New("DreamBooth empty tokens")
)

const (
	dbURL      = "https://stablediffusionapi.com/api/v4/dreambooth"
	dbFetchURL = "https://stablediffusionapi.com/api/v4/dreambooth/fetch/"
)

const (
	dbStatusProcessing = "processing"
	dbStatusError      = "error"
)

type DreamBooth struct {
	log          *zap.Logger
	retryCount   int
	retryTimeout int
	lastIDDBKey  int64
	tokens       []string
}

func NewDreamBoothAPI(log *zap.Logger, cfg *DreamBoothSettings) *DreamBooth {
	return &DreamBooth{
		log:          log,
		retryCount:   cfg.RetryCount,
		retryTimeout: cfg.RetryInterval,
		tokens:       cfg.Tokens,
	}
}

func (d *DreamBooth) GenerateText(_ context.Context, _ string) (body []byte, err error) {
	return nil, nil
}

func (d *DreamBooth) GenerateImage(ctx context.Context, prompt string) (body []byte, fileName string, err error) {
	if len(d.tokens) == 0 {
		return nil, "", errDBEmptyTokens
	}
	var retryCount int
	lastIDDBKey := atomic.LoadInt64(&d.lastIDDBKey)
	for idx := lastIDDBKey; idx < int64(len(d.tokens)); idx++ {
		retryCount++
		body, fileName, err = d.TextToImage(ctx, prompt, d.tokens[idx])
		if err == nil || !errors.Is(err, errDBMonthLimit) || retryCount == d.retryCount {
			if idx != lastIDDBKey {
				atomic.StoreInt64(&d.lastIDDBKey, idx)
			}
			return body, fileName, err
		}
		if idx == int64(len(d.tokens))-1 {
			idx = 0
		}
	}
	return nil, "", err
}

// TextToImage - https://stablediffusionapi.com/docs/community-models-api-v4/dreamboothtext2img
func (d *DreamBooth) TextToImage(ctx context.Context, text, key string) ([]byte, string, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/json")
	req.SetRequestURI(dbURL)
	reqBody := NewSerializedDBBodyRequest(key, text)
	req.SetBody(reqBody)
	d.log.Debug("DreamBooth request body:", zap.String("body", string(reqBody)))
	if err := fasthttp.Do(req, resp); err != nil {
		return nil, "", err
	}
	respBody := resp.Body()
	d.log.Debug("DreamBooth response body:", zap.String("body", string(respBody)))
	if resp.StatusCode() != fasthttp.StatusOK {
		return nil, "", errDBInvalidRespCode
	}
	return d.processResponseBody(ctx, respBody, key)
}

func (d *DreamBooth) processResponseBody(ctx context.Context, respBody []byte, key string) ([]byte, string, error) {
	var p fastjson.Parser
	v, err := p.ParseBytes(respBody)
	if err != nil {
		return nil, "", errDBParsingRespBody
	}
	outputURL := v.GetStringBytes("output", "0")
	status := string(v.GetStringBytes("status"))
	switch status {
	case dbStatusProcessing:
		requestID := strconv.Itoa(v.GetInt("id"))
		if requestID == "" {
			return nil, "", errDBRequestIDIsEmpty
		}
		outputURL, err = d.processRetryFetchQueuedImages(ctx, requestID, key)
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, "", err
		}
		if err != nil || len(outputURL) == 0 {
			return nil, "", errDBOutputIsEmpty
		}
	case dbStatusError:
		if isMonthLimitError(string(v.GetStringBytes("message"))) {
			return nil, "", errDBMonthLimit
		}
		return nil, "", errDBStatusError
	}
	return d.downloadFile(string(outputURL))
}

func (d *DreamBooth) processRetryFetchQueuedImages(ctx context.Context, requestID, key string) (outputURL []byte, err error) {
	for i := 0; i < d.retryCount; i++ {
		outputURL, err = d.FetchQueuedImages(requestID, key)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		if err != nil {
			time.Sleep(time.Duration(d.retryTimeout) * time.Second)
			continue
		}
		return
	}
	return
}

// FetchQueuedImages - https://stablediffusionapi.com/docs/community-models-api-v4/dreamboothfetchqueimg
func (d *DreamBooth) FetchQueuedImages(requestID, key string) ([]byte, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/json")
	req.SetRequestURI(dbFetchURL + requestID)
	reqBody := prepareFetchQueueImagesRequest(key, requestID)
	d.log.Debug("DreamBooth fetch request body:", zap.String("body", string(reqBody)))
	if err := fasthttp.Do(req, resp); err != nil {
		return nil, err
	}
	respBody := resp.Body()
	d.log.Debug("DreamBooth FetchQueuedImages response body:", zap.String("body", string(respBody)))
	if resp.StatusCode() != fasthttp.StatusOK {
		return nil, errDBFQIInvalidRespCode
	}
	var p fastjson.Parser
	v, err := p.ParseBytes(respBody)
	if err != nil {
		return nil, errDBFQIParsingRespBody
	}
	output := v.GetStringBytes("output", "0")
	if len(output) == 0 {
		return nil, errDBOutputIsEmpty
	}
	return output, nil
}

func (d *DreamBooth) downloadFile(fileURL string) ([]byte, string, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	req.SetRequestURI(fileURL)
	err := fasthttp.Do(req, resp)
	if err != nil {
		return nil, "", err
	}
	if resp.StatusCode() != fasthttp.StatusOK {
		return nil, "", errDBDownloadFileInvalidRespCode
	}
	respBody := resp.Body()
	if len(respBody) == 0 {
		return nil, "", errDBDownloadFileRespBodyIsEmpty
	}
	u, err := url.Parse(fileURL)
	if err != nil {
		return nil, "", err
	}
	fileName := path.Base(u.Path)
	if fileName == "" {
		return nil, "", errChatGPTEmptyFileName
	}
	return respBody, fileName, nil
}

func prepareFetchQueueImagesRequest(key, requestID string) []byte {
	var reqBody bytes.Buffer
	reqBody.WriteString(`{"key":"`)
	reqBody.WriteString(key)
	reqBody.WriteString(`","request_id":"`)
	reqBody.WriteString(requestID)
	reqBody.WriteString(`"}`)
	return reqBody.Bytes()
}

func isMonthLimitError(text string) bool {
	return strings.Contains(text, "limit exceeded")
}

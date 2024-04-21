package tbotopenai

import (
	"context"
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
	"time"

	fbAPI "github.com/dm1trypon/go-fusionbrain-api"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

const (
	defaultWidth  = 512
	defaultHeight = 512

	defaultStyle = "DEFAULT"
)

const (
	maxWidth  = 1024
	maxHeight = 1024
)

var (
	errFusionBrainEmptyModels        = errors.New("FusionBrain: empty models")
	errFusionBrainEmptyStyles        = errors.New("FusionBrain: empty styles")
	errFusionBrainInvalidRequestBody = errors.New("FusionBrain: invalid request's body")
)

type FusionBrainAPI struct {
	fb           *fbAPI.FusionBrain
	log          *zap.Logger
	retryTimeout time.Duration
}

func NewFusionBrainAPI(log *zap.Logger, cfg *FusionBrainSettings) *FusionBrainAPI {
	return &FusionBrainAPI{
		fb:           fbAPI.NewFusionBrain(&fasthttp.Client{}, cfg.Key, cfg.SecretKey),
		log:          log,
		retryTimeout: cfg.RetryInterval,
	}
}

func (f *FusionBrainAPI) GenerateText(_ context.Context, _ string) (body []byte, err error) {
	return nil, nil
}

func (f *FusionBrainAPI) GenerateImage(ctx context.Context, prompt string) (body []byte, fileName string, err error) {
	var models []fbAPI.Model
	models, err = f.fb.GetModels(ctx)
	if err != nil {
		return nil, "", err
	}
	if len(models) == 0 {
		return nil, "", errFusionBrainEmptyModels
	}
	if err = f.fb.CheckAvailable(ctx, models[0].ID); err != nil {
		return nil, "", err
	}
	var styles []fbAPI.Style
	styles, err = f.fb.GetStyles(ctx)
	if err != nil {
		return nil, "", err
	}
	if len(styles) == 0 {
		return nil, "", errFusionBrainEmptyStyles
	}
	stylesNames := make(map[string]struct{}, len(styles))
	for idx := range styles {
		stylesNames[styles[idx].Name] = struct{}{}
	}
	reqBody := validateAndPrepareFBRequestBody(prompt, stylesNames)
	if reqBody == nil {
		return nil, "", errFusionBrainInvalidRequestBody
	}
	var uuid string
	uuid, err = f.fb.TextToImage(ctx, *reqBody, models[0].ID)
	if err != nil {
		return nil, "", err
	}
	for {
		select {
		case <-ctx.Done():
			return nil, "", ctx.Err()
		default:
			var status fbAPI.GenerationStatus
			status, err = f.fb.CheckStatus(ctx, uuid)
			if err != nil {
				return nil, "", err
			}
			if status.Status == fbAPI.StatusFail {
				return nil, "", err
			}
			if status.Status == fbAPI.StatusDone {
				var imgBody []byte
				// избавляемся от кавычек с начала и с конца
				imgBodyBase64 := status.Images[0][1:][:len(status.Images[0][1:])-1]
				if imgBody, err = base64.StdEncoding.DecodeString(imgBodyBase64); err != nil {
					return nil, "", err
				}
				return imgBody, status.UUID + formatImgFile, err
			}
			time.Sleep(f.retryTimeout)
		}
	}
}

func validateAndPrepareFBRequestBody(body string, stylesNames map[string]struct{}) *fbAPI.RequestBody {
	rows := strings.Split(body, "\n")
	if len(rows) == 0 || rows[0] == "" {
		return nil
	}
	reqBody := &fbAPI.RequestBody{
		Width:  defaultWidth,
		Height: defaultHeight,
		Style:  defaultStyle,
	}
	for idx := range rows {
		switch idx {
		case 0:
			reqBody.Prompt = rows[idx]
		case 1:
			reqBody.NegativePrompt = rows[idx]
		case 2:
			width, err := strconv.Atoi(rows[idx])
			if err != nil || width <= 0 || width > maxWidth {
				continue
			}
			reqBody.Width = width
		case 3:
			height, err := strconv.Atoi(rows[idx])
			if err != nil || height <= 0 || height > maxHeight {
				continue
			}
			reqBody.Height = height
		case 4:
			style := rows[idx]
			if _, ok := stylesNames[style]; !ok {
				continue
			}
			reqBody.Style = style
		}
	}
	return reqBody
}

package tbotopenai

import (
	"bytes"
	"strconv"
	"strings"
)

// DBBodyRequest - DreamBooth Text to Image API: https://stablediffusionapi.com/docs/community-models-api-v4/dreamboothtext2img/
type DBBodyRequest struct {
	key               string
	modelID           string
	prompt            string
	negativePrompt    string
	width             string
	height            string
	samples           string
	numInferenceSteps string
	safetyChecker     string
	enhancePrompt     string
	guidanceScale     float64
	multiLingual      string
	panorama          string
	selfAttention     string
	upscale           string
	tomesd            string
	clipSkip          string
	useKarrasSigmas   string
	scheduler         string
}

func NewSerializedDBBodyRequest(key, body string) []byte {
	dbBodyReq := &DBBodyRequest{
		key:               key,
		modelID:           "midjourney",
		prompt:            "",
		negativePrompt:    "",
		width:             "1024",
		height:            "1024",
		samples:           "1",
		numInferenceSteps: "20",
		safetyChecker:     "no",
		enhancePrompt:     "yes",
		guidanceScale:     7.5,
		multiLingual:      "no",
		panorama:          "no",
		selfAttention:     "no",
		upscale:           "no",
		tomesd:            "yes",
		clipSkip:          "2",
		useKarrasSigmas:   "yes",
		scheduler:         "UniPCMultistepScheduler",
	}
	dbBodyReq.fillChangedFields(body)
	return dbBodyReq.serialize()
}

func (d *DBBodyRequest) fillChangedFields(body string) {
	body = strings.ReplaceAll(body, "\r", "")
	body = strings.Replace(body, " ", "", 1)
	parts := strings.Split(body, "\n")
	for i := 0; i < len(parts); i++ {
		field, val, found := strings.Cut(parts[i], ":")
		if !found || field == "" || val == "" {
			continue
		}
		switch field {
		case "model_id":
			d.modelID = val
		case "prompt":
			d.prompt = val
		case "enhance_prompt":
			d.enhancePrompt = val
		case "negative_prompt":
			d.negativePrompt = val
		case "width":
			d.width = val
		case "height":
			d.height = val
		case "samples":
			d.samples = val
		case "num_inference_steps":
			d.numInferenceSteps = val
		case "safety_checker":
			d.safetyChecker = val
		case "guidance_scale":
			guidanceScale, err := strconv.ParseFloat(val, 64)
			if err != nil {
				break
			}
			d.guidanceScale = guidanceScale
		case "multi_lingual":
			d.multiLingual = val
		case "panorama":
			d.panorama = val
		case "self_attention":
			d.selfAttention = val
		case "upscale":
			d.upscale = val
		case "tomesd":
			d.tomesd = val
		case "clip_skip":
			d.clipSkip = val
		case "use_karras_sigmas":
			d.useKarrasSigmas = val
		case "scheduler":
			d.scheduler = val
		}
	}
}

func (d *DBBodyRequest) serialize() []byte {
	var b bytes.Buffer
	b.WriteString(`{"key":"`)
	b.WriteString(d.key)
	b.WriteString(`","model_id":"`)
	b.WriteString(d.modelID)
	b.WriteString(`","prompt":"`)
	b.WriteString(d.prompt)
	b.WriteString(`","negative_prompt":"`)
	b.WriteString(d.negativePrompt)
	b.WriteString(`","width":"`)
	b.WriteString(d.width)
	b.WriteString(`","height":"`)
	b.WriteString(d.height)
	b.WriteString(`","samples":"`)
	b.WriteString(d.samples)
	b.WriteString(`","num_inference_steps":"`)
	b.WriteString(d.negativePrompt)
	b.WriteString(`","safety_checker":"`)
	b.WriteString(d.safetyChecker)
	b.WriteString(`","enhance_prompt":"`)
	b.WriteString(d.enhancePrompt)
	b.WriteString(`","guidance_scale":`)
	b.WriteString(strconv.FormatFloat(d.guidanceScale, 'f', 2, 64))
	b.WriteString(`,"multi_lingual":"`)
	b.WriteString(d.multiLingual)
	b.WriteString(`","panorama":"`)
	b.WriteString(d.panorama)
	b.WriteString(`","self_attention":"`)
	b.WriteString(d.selfAttention)
	b.WriteString(`","upscale":"`)
	b.WriteString(d.upscale)
	b.WriteString(`","tomesd":"`)
	b.WriteString(d.tomesd)
	b.WriteString(`","clip_skip":"`)
	b.WriteString(d.clipSkip)
	b.WriteString(`","use_karras_sigmas":"`)
	b.WriteString(d.useKarrasSigmas)
	b.WriteString(`","scheduler":"`)
	b.WriteString(d.scheduler)
	b.WriteString(`"}`)
	return b.Bytes()
}

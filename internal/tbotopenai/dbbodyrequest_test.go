package tbotopenai

//
//func TestDBBodyRequest_MarshalJSON(t *testing.T) {
//	req := newDBBodyRequestTest()
//	tests := []struct {
//		name      string
//		expResult []byte
//		expError  error
//	}{
//		{
//			name:      "Success marshaling JSON",
//			expResult: []byte(`{"key":"key","model_id":"modelID","prompt":"prompt","negative_prompt":"negativePrompt","width":"width","height":"height","samples":"samples","num_inference_steps":"negativePrompt","safety_checker":"safetyChecker","enhance_prompt":"enhancePrompt","guidance_scale":1.01,"multi_lingual":"multiLingual","panorama":"panorama","self_attention":"selfAttention","upscale":"upscale","tomesd":"tomesd","clip_skip":"clipSkip","use_karras_sigmas":"useKarrasSigmas","scheduler":"scheduler"}`),
//			expError:  nil,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			body, err := req.MarshalJSON()
//			assert.Equal(t, string(tt.expResult), string(body))
//			assert.Equal(t, tt.expError, err)
//		})
//	}
//}
//
//func TestDBBodyRequest_FillChangedFields(t *testing.T) {
//	req := newDBBodyRequestTest()
//	tests := []struct {
//		name             string
//		body             string
//		expDBBodyRequest *DBBodyRequest
//	}{
//		{
//			name: "Success",
//			body: "model_id:test_model_id\n\rprompt: test prompt\n\renhance_prompt:test_enhance_prompt\nnegative_prompt:test_negative_prompt\r\nwidth:100\n\rheight:200",
//			expDBBodyRequest: &DBBodyRequest{
//				modelID:        "test_model_id",
//				prompt:         "test prompt",
//				enhancePrompt:  "test_enhance_prompt",
//				negativePrompt: "test_negative_prompt",
//				width:          "100",
//				height:         "200",
//			},
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			req.FillChangedFields(tt.body)
//			assert.Equal(t, tt.expDBBodyRequest.modelID, req.modelID)
//			assert.Equal(t, tt.expDBBodyRequest.prompt, req.prompt)
//			assert.Equal(t, tt.expDBBodyRequest.enhancePrompt, req.enhancePrompt)
//			assert.Equal(t, tt.expDBBodyRequest.negativePrompt, req.negativePrompt)
//			assert.Equal(t, tt.expDBBodyRequest.width, req.width)
//			assert.Equal(t, tt.expDBBodyRequest.height, req.height)
//		})
//	}
//}
//
//func newDBBodyRequestTest() *DBBodyRequest {
//	return &DBBodyRequest{
//		key:               "key",
//		modelID:           "modelID",
//		prompt:            "prompt",
//		negativePrompt:    "negativePrompt",
//		width:             "width",
//		height:            "height",
//		samples:           "samples",
//		numInferenceSteps: "numInferenceSteps",
//		safetyChecker:     "safetyChecker",
//		enhancePrompt:     "enhancePrompt",
//		guidanceScale:     1.01,
//		multiLingual:      "multiLingual",
//		panorama:          "panorama",
//		selfAttention:     "selfAttention",
//		upscale:           "upscale",
//		tomesd:            "tomesd",
//		clipSkip:          "clipSkip",
//		useKarrasSigmas:   "useKarrasSigmas",
//		scheduler:         "scheduler",
//	}
//}

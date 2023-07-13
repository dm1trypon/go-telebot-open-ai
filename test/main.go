package main

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"log"
)

const url = "https://stablediffusionapi.com/api/v4/dreambooth"
const urlFetch = "https://stablediffusionapi.com/api/v4/dreambooth/fetch"

// 6339322764:AAGXPnK3BDqYKRuvXP6JUghl4ffh5xkaV4A
// 5930839504:AAEKsOufyhnQuwL3kOTJJQyHzJnJqAWY0GU
const body = `{
  "key": "L9bzBIoeETHHEfhEHDxIBpQRlGQEMb1BkDDxCzpm2SIgI4Fkqa36X8mqU3NK",
  "model_id": "midjourney",
  "prompt": "ultra realistic close up portrait ((beautiful pale cyberpunk female with heavy black eyeliner)), blue eyes, shaved side haircut, hyper detail, cinematic lighting, magic neon, dark red city, Canon EOS R3, nikon, f/1.4, ISO 200, 1/160s, 8K, RAW, unedited, symmetrical balance, in-frame, 8K",
  "negative_prompt": "painting, extra fingers, mutated hands, poorly drawn hands, poorly drawn face, deformed, ugly, blurry, bad anatomy, bad proportions, extra limbs, cloned face, skinny, glitchy, double torso, extra arms, extra hands, mangled fingers, missing lips, ugly face, distorted face, extra legs, anime",
  "width": "512",
  "height": "512",
  "samples": "1",
  "num_inference_steps": "20",
  "safety_checker": "no",
  "enhance_prompt": "yes",
  "seed": null,
  "guidance_scale": 7.5,
  "multi_lingual": "no",
  "panorama": "no",
  "self_attention": "no",
  "upscale": "no",
  "embeddings_model": null,
  "lora_model": null,
  "clip_skip": "2",
  "tomesd": "yes",
  "use_karras_sigmas": "yes",
  "vae": null,
  "lora_strength": null,
  "scheduler": "UniPCMultistepScheduler",
  "webhook": null,
  "track_id": null
}`

const bodyFetch = `{
 "key": "kD1tMhkcKz5p1wZlgHXWcfiZrAp5OUXfBPBjlHz2DFX8xNBGL6r2wIndvTdP",
 "request_id": "27354295"
}`

const bodyTwo = "{\"key\":\"ZJ3ijZcL9CCGcYFdSe5PP0sFv0Pi7LJYDZ7KdaAqKRnUX7amilcy7nKf5IkE\",\"model_id\":\"runwayml/stable-diffusion-v1-5\",\"prompt\":\"preteen beautiful girl, photorealistic, 4K\",\"negative_prompt\":\"\",\"width\":\"1024\",\"height\":\"1024\",\"samples\":\"1\",\"num_inference_steps\":\"\",\"safety_checker\":\"no\",\"enhance_prompt\":\"yes\",\"guidance_scale\":7.50,\"multi_lingual\":\"no\",\"panorama\":\"no\",\"self_attention\":\"no\",\"upscale\":\"no\",\"tomesd\":\"yes\",\"clip_skip\":\"2\",\"use_karras_sigmas\":\"yes\",\"scheduler\":\"UniPCMultistepScheduler\"}"

func main() {
	//url := "https://stablediffusionapi.com/api/v3/text2img"

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod("POST")
	req.Header.SetContentType("application/json")
	req.SetRequestURI(url)
	req.SetBody([]byte(body))

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Отправьте POST-запрос
	if err := fasthttp.Do(req, resp); err != nil {
		log.Fatal("Ошибка при выполнении запроса:", err)
	}

	// Получите ответ
	responseBody := resp.Body()

	fmt.Println("Статус код:", resp.StatusCode())
	fmt.Println("Ответ:", string(responseBody))
}

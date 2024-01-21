package tbotopenai

import (
	"bytes"
	"errors"
	"strconv"
	"strings"
)

const (
	respBodySessionIsAlreadyExist = `⚠ Сессия с ботом уже активна ⚠ 
🔧 /help - описание команд`
	respBodySessionCreated = `🎉 Сессия активна, добро пожаловать! 🎉
🔧 /help - описание команд`
	respBodySessionRemoved    = `😥 Сессия с ботом завершена. Возвращайтесь! 😥`
	respBodySessionIsNotExist = `❌ Сессия с ботом не активна ❌
✅ /start - начало сессии с ботом
🔧 /help - описание команд`
	respBodyCommandChatGPT = `📖 Генерация текста с помощью ChatGPT, модель gpt-3.5-turbo-1106 📖
Введите запрос как можно подробнее, чтобы получить наиболее удовлетворительный сгенерированный текстовый ответ`
	respBodyCommandOpenAIText = `📖 Генерация текста с помощью OpenAI, модель gpt-4-32k-0613 📖
Введите запрос как можно подробнее, чтобы получить наиболее удовлетворительный сгенерированный текстовый ответ`
	respBodyCommandOpenAIImage = `🌄 Генерация изображений с помощью OpenAI 🌄
Введите запрос как можно подробнее, чтобы получить наиболее удовлетворительное сгенерированное изображение`
	respBodyCommandDreamBooth = `🌅 Выбрана генерация изображений с помощью DreamBooth 🌅
⚠ Для лучшего результата ознакомтесь с документацией https://stablediffusionapi.com/docs/community-models-api-v4/dreamboothtext2img#body-attributes ⚠
📄 /dreamBoothExample - пример промпта для генерации изображения через API DreamBooth`
	respBodyUndefinedGeneration = `❌ Не выбран AI для генерации ❌
📖 /chatGPT - генерация текста, используя API ресурса chatgptbot.ru (Модель gpt-3.5-turbo-1106)
📖 /openAIText - генерация текста, используя API OpenAI (Модель gpt-4-32k-0613)
🌄 /openAIImage - генерация изображения размером 1024x1024, используя API OpenAI
🌅 /dreamBooth - продвинутая генерация изображений, используя API DreamBooth
📄 /dreamBoothExample - пример промпта для генерации изображения через API DreamBooth
🔧 /help - описание команд`
	respBodyUndefinedCommand = `❌ Комманда не поддерживается ❌
Чтобы посмотреть описание команд, введите команду /help`
	respBodyCommandDreamBoothExample = `prompt: Iron Man, (Arnold Tsang, Toru Nakayama), Masterpiece, Studio Quality, 6k , toa, toaair, 1boy, glowing, axe, mecha, science_fiction, solo, weapon, jungle , green_background, nature, outdoors, solo, tree, weapon, mask, dynamic lighting, detailed shading, digital texture painting
negative_prompt: un-detailed skin, semi-realistic, cgi, 3d, render, sketch, cartoon, drawing, ugly eyes, (out of frame:1.3), worst quality, low quality, jpeg artifacts, cgi, sketch, cartoon, drawing, (out of frame:1.1)
width: 512
height: 512
model_id: midjourney`
	respBodyInputJobID = `📛 Введите номер запроса 📛
📋 /listJobs - список выполняющихся запросов в очереди`
	respBodyRequestAddedToQueue = `✅ Запрос добавлен в очередь ✅`
	respErrBodyLimitMessages    = `❌ Сервис перегружен запросами ❌
Пожалуйста, выполните запрос позже`
	respErrBodyLimitJobs = `❌ Превышен лимит запросов ❌
Пожалуйста, дождитесь выполнения прошлых и повторите`
	respErrBodyInvalidFormatJobID = `❌ Номер задачи должен быть числом ❌`
	respErrBodyChatGPT            = `❌ Произошла ошибка при генерации ответа ChatGPT ❌
Попробуйте еще раз`
	respErrBodyOpenAI = `❌ Произошла ошибка при генерации ответа OpenAI ❌
Попробуйте еще раз`
	respErrBodyDreamBoothByStatusCode = `❌ Произошла ошибка при генерации ответа DreamBooth ❌
К сожалению, в данный момент сервис DreamBooth не работает, попробуйте выполнить запрос позже`
	respErrBodyDreamBooth = `❌ Произошла ошибка при генерации изображения DreamBooth ❌
Попробуйте еще раз`
	respErrBodyJobCanceled = `✅ Запрос был отменен ✅`
)

func respErrBodyCommandDreamBooth(err error) string {
	if errors.Is(err, errDBInvalidRespCode) {
		return respErrBodyDreamBoothByStatusCode
	}
	return respErrBodyDreamBooth
}

func respErrBodyJobIsNotExist(jobID int) []byte {
	var b bytes.Buffer
	b.WriteString("Задача №")
	b.WriteString(strconv.Itoa(jobID))
	b.WriteString(" не найдена")
	return b.Bytes()
}

func respBodySuccessCancelJob(api string, jobID int) []byte {
	var b bytes.Buffer
	b.WriteString("Задача ")
	b.WriteString(api)
	b.WriteString(" №")
	b.WriteString(strconv.Itoa(jobID))
	b.WriteString(" завершена")
	return b.Bytes()
}

func respBodyListJobs(textJobIDs, imgJobIDs, openAIIDs []int) string {
	var b strings.Builder
	b.WriteString("Список задач ChatGPT:\r\n")
	for i := range textJobIDs {
		b.WriteString(strconv.Itoa(textJobIDs[i]))
		b.WriteString("\r\n")
	}
	b.WriteString("Список задач DreamBooth:\r\n")
	for i := range imgJobIDs {
		b.WriteString(strconv.Itoa(imgJobIDs[i]))
		b.WriteString("\r\n")
	}
	b.WriteString("Список задач OpenAI:\r\n")
	for i := range openAIIDs {
		b.WriteString(strconv.Itoa(openAIIDs[i]))
		b.WriteString("\r\n")
	}
	return b.String()
}

func respBodyCommandHelp(role string) string {
	var b strings.Builder
	b.WriteString(`🔧 Доступные команды бота 🔧
✅ /start - начало сессии с ботом
⛔ /stop - завершение сессии с ботом
📖 /chatGPT - генерация текста, используя API ресурса chatgptbot.ru (Модель gpt-3.5-turbo-1106)
`)
	if role == roleAdmin {
		b.WriteString(`📖 /openAIText - генерация текста, используя API OpenAI (Модель gpt-4-32k-0613)
🌄 /openAIImage - генерация изображения размером 1024x1024, используя API OpenAI
`)
	}
	b.WriteString(`🌅 /dreamBooth - продвинутая генерация изображений, используя API DreamBooth
📄 /dreamBoothExample - пример промпта для генерации изображения через API DreamBooth
📛 /cancelJob - отмена текущего запроса по ее номеру
📋 /listJobs - список выполняющихся запросов в очереди`)
	return b.String()
}

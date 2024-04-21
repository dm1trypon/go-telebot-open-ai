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
⚠ Для лучшего результата ознакомьтесь с документацией https://stablediffusionapi.com/docs/community-models-api-v4/dreamboothtext2img#body-attributes ⚠
📄 /dreamBoothExample - пример промпта для генерации изображения через API DreamBooth`
	respBodyUndefinedJob = `❌ Не выбрана команда для выполнения задачи ❌
🔧 /help - описание команд`
	respBodyUndefinedCommand = `❌ Комманда не поддерживается ❌
Чтобы посмотреть описание команд, введите команду /help`
	respBodyAccessDenied             = `❌ Доступ запрещен ❌`
	respBodyCommandDreamBoothExample = `prompt: Iron Man, (Arnold Tsang, Toru Nakayama), Masterpiece, Studio Quality, 6k , toa, toaair, 1boy, glowing, axe, mecha, science_fiction, solo, weapon, jungle , green_background, nature, outdoors, solo, tree, weapon, mask, dynamic lighting, detailed shading, digital texture painting
negative_prompt: un-detailed skin, semi-realistic, cgi, 3d, render, sketch, cartoon, drawing, ugly eyes, (out of frame:1.3), worst quality, low quality, jpeg artifacts, cgi, sketch, cartoon, drawing, (out of frame:1.1)
width: 512
height: 512
model_id: midjourney`
	respBodyInputJobID = `📛 Введите номер запроса 📛
📋 /listJobs - список выполняющихся запросов в очереди`
	respBodyRequestAddedToQueue = `✅ Запрос добавлен в очередь ✅`
	respBodyStatsCommand        = `☣ Здесь должен быть файл со статистикой запросов и ответов ☣`
	respBodyCommandBan          = `🌄 Введите имя пользователя для бана 🌄`
	respBodyCommandUnban        = `🌄 Введите имя пользователя для разбана 🌄`
	respBodyRequestBan          = `✅ Пользователь забанен ✅`
	respBodyRequestUnban        = `✅ Пользователь разбанен ✅`
	respBodyCommandFusionBrain  = `🌅 Выбрана генерация изображений с помощью FusionBrain 🌅
📄 /fusionBrainExample - пример запроса для генерации изображения через API FusionBrain`
	respBodyCommandFusionBrainExample = `Порядок заполнения запроса для генерации изабражения с помощью Fusion Brain API:
1 - (обязательно для заполнения) какие цвета и приёмы модель должна использовать при генерации изображения
2 - какие цвета и приёмы модель не должна использовать при генерации изображения
3 - (если не задано, по-умолчанию будет 512) длина изображения
4 - (если не задано, по-умолчанию будет 512) ширина изображения
5 - (если не задано, по-умолчанию будет DEFAULT) стиль изображения:
	- KANDINSKY https://cdn.fusionbrain.ai/static/download/img-style-kandinsky.png
	- UHD https://cdn.fusionbrain.ai/static/download/img-style-detail-photo.png
	- ANIME https://cdn.fusionbrain.ai/static/download/img-style-anime.png
	- DEFAULT https://cdn.fusionbrain.ai/static/download/img-style-personal.png
Если какой-то параметр не нужен, оставляем строку пустой и переходим на следующую строку, например:
Пушистый кот в очках

1024
1024
DEFAULT
- вторая строка пустая, или:
Пушистый кот в очках
яркие цвета, кислотность, высокая контрастность
1024
1024

- строка после ширины пустая, будет использован стиль по-умолчанию - DEFAULT`
	respErrBodyRequestBan                     = `❌ Произошла ошибка при бане пользователя ❌`
	respErrBodyRequestUnban                   = `❌ Произошла ошибка при разбане пользователя ❌`
	respErrBodyRequestUnbanUsernameIsNotExist = `❌ Пользователя нет в черном списке ❌`
	respErrBodyRequestBanUsernameAlreadyExist = `❌ Пользователь уже есть в черном списке ❌`
	respErrBodyLimitMessages                  = `❌ Сервис перегружен запросами ❌
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
	respErrBodyFusionBrain = `❌ Произошла ошибка при генерации изображения FusionBrain ❌
Попробуйте еще раз`
	respErrBodyJobCanceled = `✅ Запрос был отменен ✅`
	respErrBodyGetLogs     = `❌ Произошла ошибка при получении логов ❌`
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
	b.WriteString(" завершена.\n")
	b.WriteString(respBodyInputJobID)
	return b.Bytes()
}

func respBodyListJobs(textJobIDs, imgJobIDs, openAIIDs, fusionBrainIDs []int, role string) string {
	var b strings.Builder
	b.WriteString("Список задач ChatGPT:\r\n")
	for i := range textJobIDs {
		b.WriteString(strconv.Itoa(textJobIDs[i]))
		b.WriteString("\r\n")
	}
	if role == roleAdmin {
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
	}
	b.WriteString("Список задач FusionBrain:\r\n")
	for i := range fusionBrainIDs {
		b.WriteString(strconv.Itoa(fusionBrainIDs[i]))
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
🌅 /fusionBrain - продвинутая генерация изображений, используя API FusionBrain
📄 /fusionBrainExample - пример промпта для генерации изображения через API FusionBrain
`)
	if role == roleAdmin {
		b.WriteString(`📖 /openAIText - генерация текста, используя API OpenAI (Модель gpt-4-32k-0613)
🌄 /openAIImage - генерация изображения размером 1024x1024, используя API OpenAI
🌅 /dreamBooth - продвинутая генерация изображений, используя API DreamBooth
📄 /dreamBoothExample - пример промпта для генерации изображения через API DreamBooth
`)
	}
	b.WriteString(`📛 /cancelJob - отмена текущего запроса по ее номеру
📋 /listJobs - список выполняющихся запросов в очереди
`)
	if role == roleAdmin {
		b.WriteString(`📈 /stats - статистика запросов и ответов всех пользователей в формате csv
💻 /logs - логи сервиса
👎 /ban - бан пользователя
👍 /unban - разбан пользователя
💩 /blacklist - список заблокированных пользователей
`)
	}
	return b.String()
}

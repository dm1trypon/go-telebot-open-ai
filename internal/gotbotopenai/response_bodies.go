package gotbotopenai

import (
	"errors"
	"strings"
)

const (
	respBodySessionAlreadyExist = "Сессия с ботом уже активна. Чтобы посмотреть описание команд, введите команду /help."
	respBodySessionCreated      = "Сессия с ботом активна. Чтобы посмотреть описание команд, введите команду /help."
	respBodySessionRemoved      = "Сессия с ботом завершена. Возвращайтесь!"
	respBodySessionIsNotExist   = "Сессия с ботом не активна. Чтобы начать сессию с ботом, введите команду /start. Чтобы посмотреть описание команд, введите команду /help."
	respBodyLimitJobs           = "Прошлые запросы еще не обработаны. Пожалуйста, дождитесь их выполнения и повторите запрос."
	respBodyCommandText         = "Выбрана генерация текста. Введите запрос как можно подробнее, чтобы получить наиболее удовлетворительный сгенерированный текстовый ответ."
	respBodyTmplCommandImage    = "Выбрана генерация изображений размером %s. Введите запрос как можно подробнее, чтобы получить наиболее удовлетворительное сгенерированное изображение."
	respBodyCommandImageCustom  = "Выбрана пользовательская генерация изображений. Введите запрос как можно подробнее, чтобы получить наиболее удовлетворительное сгенерированное изображение."
	respBodyUndefinedGeneration = "Не выбрано, что генерировать: текст или изображение. Чтобы посмотреть описание команд, введите команду /help."
	respBodyUndefinedCommand    = "Комманда не поддерживается. Чтобы посмотреть описание команд, введите команду /help."
	respBodyCommandHelp         = `Доступные команды бота:
/start - начало сессии с ботом
/stop - завершение сессии с ботом
/image256x256 - генерация изображений размером 256x256, используя модель OpenAI DALL·E
/image512x512 - генерация изображений размером 512x512, используя модель OpenAI DALL·E
/image1024x1024 - генерация изображений размером 1024x1024, используя модель OpenAI DALL·E
/imageCustom - генерация изображений, используя разные модели и более точные настройки. Описание доступных полей https://stablediffusionapi.com/docs/community-models-api-v4/dreamboothtext2img#body-attributes.
/imageCustomExample - пример промта для генерации изображения с кастомными моделями и настройками.
/text - генерация текста, используя модель OpenAI gpt-4-32k-0613.`
	respBodyCommandImageCustomExample = `prompt: Iron Man, (Arnold Tsang, Toru Nakayama), Masterpiece, Studio Quality, 6k , toa, toaair, 1boy, glowing, axe, mecha, science_fiction, solo, weapon, jungle , green_background, nature, outdoors, solo, tree, weapon, mask, dynamic lighting, detailed shading, digital texture painting
negative_prompt: un-detailed skin, semi-realistic, cgi, 3d, render, sketch, cartoon, drawing, ugly eyes, (out of frame:1.3), worst quality, low quality, jpeg artifacts, cgi, sketch, cartoon, drawing, (out of frame:1.1)
width: 512
height: 512
model_id: midjourney`

	respErrBodyCommandText                    = "Запрос не удовлетворяет политике работы с текстами OpenAI https://openai.com/policies/usage-policies. Пожалуйста, переформулируйте запрос."
	respErrBodyCommandImage                   = "Запрос не удовлетворяет политике работы с изображениями OpenAI https://openai.com/policies/usage-policies. Пожалуйста, переформулируйте запрос."
	respErrBodyCommandImageCustomByStatusCode = "К сожалению, в данный момент сервис пользовательской генерации изображений не работает, попробуйте позже."
	respErrBodyCommandImageCustomByPolitic    = "Увы, но изображение так и не сгенерировалось. Возможно, некоторые параметры были подобраны неправильно, либо запрос не удовлетворяет политике работы с OpenAI https://openai.com/policies/usage-policies, или же сервис попросту перегружен. Попробуйте переформулировать запрос и отправить его чуть позже."
)

func respBodyCommandImage(size string) string {
	return strings.Replace(respBodyTmplCommandImage, "%s", size, 1)
}

func respErrBodyCommandImageCustom(err error) string {
	if errors.Is(err, errDBInvalidRespCode) {
		return respErrBodyCommandImageCustomByStatusCode
	}
	return respErrBodyCommandImageCustomByPolitic
}

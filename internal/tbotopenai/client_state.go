package tbotopenai

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var (
	errChatGPTJobIsNotExist       = errors.New("ChatGPT job '%d' is not exist")
	errDreamBoothJobIsNotExist    = errors.New("DreamBooth job '%d' is not exist")
	errOpenAIJobIsNotExist        = errors.New("OpenAI job '%d' is not exist")
	errChatGPTJobIsAlreadyUsed    = errors.New("ChatGPT job '%d' is already used")
	errOpenAIJobIsAlreadyUsed     = errors.New("OpenAI job '%d' is already used")
	errDreamBoothJobIsAlreadyUsed = errors.New("DreamBooth job '%d' is already used")
	chatIDIsNotExistErr           = errors.New("client with current chatID is not exist")
	chatIDAlreadyExistErr         = errors.New("client with current chatID already exist")
)

func ErrorChatGPTJobIsNotExist(id int) error {
	return fmt.Errorf(errChatGPTJobIsNotExist.Error(), id)
}

func ErrorDreamBoothJobIsNotExist(id int) error {
	return fmt.Errorf(errDreamBoothJobIsNotExist.Error(), id)
}

func ErrorOpenAIJobIsNotExist(id int) error {
	return fmt.Errorf(errOpenAIJobIsNotExist.Error(), id)
}

func ErrorChatGPTJobIsAlreadyUsed(id int) error {
	return fmt.Errorf(errChatGPTJobIsAlreadyUsed.Error(), id)
}

func ErrorOpenAIJobIsAlreadyUsed(id int) error {
	return fmt.Errorf(errOpenAIJobIsAlreadyUsed.Error(), id)
}

func ErrorDreamBoothJobIsAlreadyUsed(id int) error {
	return fmt.Errorf(errDreamBoothJobIsAlreadyUsed.Error(), id)
}

type clientState struct {
	command        string
	openAICancels  map[int]context.CancelFunc
	chatGPTCancels map[int]context.CancelFunc
	ddCancels      map[int]context.CancelFunc
}

func NewTClient() *clientState {
	return &clientState{
		command:        commandStart,
		openAICancels:  make(map[int]context.CancelFunc),
		chatGPTCancels: make(map[int]context.CancelFunc),
		ddCancels:      make(map[int]context.CancelFunc),
	}
}

func (c *clientState) LenOpenAIJobs() int {
	return len(c.openAICancels)
}

func (c *clientState) LenChatGPTJobs() int {
	return len(c.chatGPTCancels)
}

func (c *clientState) LenDreamBoothJobs() int {
	return len(c.ddCancels)
}

func (c *clientState) OpenAIJobs() []int {
	jobIDs := make([]int, 0, len(c.openAICancels))
	for id := range c.openAICancels {
		jobIDs = append(jobIDs, id)
	}
	return jobIDs
}

func (c *clientState) ChatGPTJobs() []int {
	jobIDs := make([]int, 0, len(c.chatGPTCancels))
	for id := range c.chatGPTCancels {
		jobIDs = append(jobIDs, id)
	}
	return jobIDs
}

func (c *clientState) DreamBoothJobs() []int {
	jobIDs := make([]int, 0, len(c.chatGPTCancels))
	for id := range c.ddCancels {
		jobIDs = append(jobIDs, id)
	}
	return jobIDs
}

func (c *clientState) SetCommand(command string) {
	c.command = command
}

func (c *clientState) Command() string {
	return c.command
}

func (c *clientState) SetCancelOpenAIJob(cancel context.CancelFunc, id int) error {
	if _, ok := c.openAICancels[id]; ok {
		return ErrorOpenAIJobIsAlreadyUsed(id)
	}
	c.openAICancels[id] = cancel
	return nil
}

func (c *clientState) SetCancelChatGPTJob(cancel context.CancelFunc, id int) error {
	if _, ok := c.chatGPTCancels[id]; ok {
		return ErrorChatGPTJobIsAlreadyUsed(id)
	}
	c.chatGPTCancels[id] = cancel
	return nil
}

func (c *clientState) SetCancelDreamBoothJob(cancel context.CancelFunc, id int) error {
	if _, ok := c.ddCancels[id]; ok {
		return ErrorDreamBoothJobIsAlreadyUsed(id)
	}
	c.ddCancels[id] = cancel
	return nil
}

func (c *clientState) CancelChatGPTJob(id int) error {
	cancel, ok := c.chatGPTCancels[id]
	if !ok {
		return ErrorChatGPTJobIsNotExist(id)
	}
	cancel()
	delete(c.chatGPTCancels, id)
	return nil
}

func (c *clientState) CancelDreamBoothJob(id int) error {
	cancel, ok := c.ddCancels[id]
	if !ok {
		return ErrorDreamBoothJobIsNotExist(id)
	}
	cancel()
	delete(c.ddCancels, id)
	return nil
}

func (c *clientState) CancelOpenAIJob(id int) error {
	cancel, ok := c.openAICancels[id]
	if !ok {
		return ErrorOpenAIJobIsNotExist(id)
	}
	cancel()
	delete(c.openAICancels, id)
	return nil
}

func (c *clientState) CancelChatGPTJobs() {
	for _, cancel := range c.chatGPTCancels {
		cancel()
	}
	c.chatGPTCancels = make(map[int]context.CancelFunc)
}

func (c *clientState) CancelDreamBoothJobs() {
	for _, cancel := range c.ddCancels {
		cancel()
	}
	c.ddCancels = make(map[int]context.CancelFunc)
}

func (c *clientState) CancelOpenAIJobs() {
	for _, cancel := range c.openAICancels {
		cancel()
	}
	c.openAICancels = make(map[int]context.CancelFunc)
}

type clientStateByChatID struct {
	value map[int64]*clientState
	mutex sync.RWMutex
}

func (t *clientStateByChatID) AddClient(chatID int64) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	_, ok := t.value[chatID]
	if ok {
		return chatIDAlreadyExistErr
	}
	t.value[chatID] = NewTClient()
	return nil
}

func (t *clientStateByChatID) DeleteClient(chatID int64) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	tc, ok := t.value[chatID]
	if !ok || tc == nil {
		return chatIDIsNotExistErr
	}
	delete(t.value, chatID)
	return nil
}

func (t *clientStateByChatID) UpdateClientCommand(chatID int64, command string) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	tc, ok := t.value[chatID]
	if !ok || tc == nil {
		return chatIDIsNotExistErr
	}
	tc.SetCommand(command)
	return nil
}

func (t *clientStateByChatID) ClientCommand(chatID int64) (string, error) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	tc, ok := t.value[chatID]
	if !ok || tc == nil {
		return "", chatIDIsNotExistErr
	}
	return tc.Command(), nil
}

func (t *clientStateByChatID) ClientChatGPTJobs(chatID int64) ([]int, error) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	tc, ok := t.value[chatID]
	if !ok || tc == nil {
		return nil, chatIDIsNotExistErr
	}
	return tc.ChatGPTJobs(), nil
}

func (t *clientStateByChatID) ClientDreamBoothJobs(chatID int64) ([]int, error) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	tc, ok := t.value[chatID]
	if !ok || tc == nil {
		return nil, chatIDIsNotExistErr
	}
	return tc.DreamBoothJobs(), nil
}

func (t *clientStateByChatID) ClientOpenAIJobs(chatID int64) ([]int, error) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	tc, ok := t.value[chatID]
	if !ok || tc == nil {
		return nil, chatIDIsNotExistErr
	}
	return tc.OpenAIJobs(), nil
}

func (t *clientStateByChatID) ClientLenChatGPTJobs(chatID int64) (int, error) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	tc, ok := t.value[chatID]
	if !ok || tc == nil {
		return -1, chatIDIsNotExistErr
	}
	return tc.LenChatGPTJobs(), nil
}

func (t *clientStateByChatID) ClientLenDreamBoothJobs(chatID int64) (int, error) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	tc, ok := t.value[chatID]
	if !ok || tc == nil {
		return -1, chatIDIsNotExistErr
	}
	return tc.LenDreamBoothJobs(), nil
}

func (t *clientStateByChatID) ClientLenOpenAIJobs(chatID int64) (int, error) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	tc, ok := t.value[chatID]
	if !ok || tc == nil {
		return -1, chatIDIsNotExistErr
	}
	return tc.LenOpenAIJobs(), nil
}

func (t *clientStateByChatID) ClientAddChatGPTJob(cancel context.CancelFunc, jobID int, chatID int64) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	tc, ok := t.value[chatID]
	if !ok || tc == nil {
		return chatIDIsNotExistErr
	}
	return tc.SetCancelChatGPTJob(cancel, jobID)
}

func (t *clientStateByChatID) ClientAddDreamBoothJob(cancel context.CancelFunc, jobID int, chatID int64) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	tc, ok := t.value[chatID]
	if !ok || tc == nil {
		return chatIDIsNotExistErr
	}
	return tc.SetCancelDreamBoothJob(cancel, jobID)
}

func (t *clientStateByChatID) ClientAddOpenAIJob(cancel context.CancelFunc, jobID int, chatID int64) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	tc, ok := t.value[chatID]
	if !ok || tc == nil {
		return chatIDIsNotExistErr
	}
	return tc.SetCancelOpenAIJob(cancel, jobID)
}

func (t *clientStateByChatID) ClientCancelChatGPTJob(jobID int, chatID int64) error {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	tc, ok := t.value[chatID]
	if !ok || tc == nil {
		return chatIDIsNotExistErr
	}
	return tc.CancelChatGPTJob(jobID)
}

func (t *clientStateByChatID) ClientCancelDreamBoothJob(jobID int, chatID int64) error {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	tc, ok := t.value[chatID]
	if !ok || tc == nil {
		return chatIDIsNotExistErr
	}
	return tc.CancelDreamBoothJob(jobID)
}

func (t *clientStateByChatID) ClientCancelOpenAIJob(jobID int, chatID int64) error {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	tc, ok := t.value[chatID]
	if !ok || tc == nil {
		return chatIDIsNotExistErr
	}
	return tc.CancelOpenAIJob(jobID)
}

func (t *clientStateByChatID) ClientCancelJobs(chatID int64) error {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	tc, ok := t.value[chatID]
	if !ok || tc == nil {
		return chatIDIsNotExistErr
	}
	tc.CancelChatGPTJobs()
	tc.CancelDreamBoothJobs()
	return nil
}

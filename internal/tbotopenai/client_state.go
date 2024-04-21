package tbotopenai

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var (
	errChatGPTJobIsNotExist        = errors.New("ChatGPT job '%d' is not exist")
	errDreamBoothJobIsNotExist     = errors.New("DreamBooth job '%d' is not exist")
	errOpenAIJobIsNotExist         = errors.New("OpenAI job '%d' is not exist")
	errFusionBrainJobIsNotExist    = errors.New("FusionBrain job '%d' is not exist")
	errChatGPTJobIsAlreadyUsed     = errors.New("ChatGPT job '%d' is already used")
	errOpenAIJobIsAlreadyUsed      = errors.New("OpenAI job '%d' is already used")
	errDreamBoothJobIsAlreadyUsed  = errors.New("DreamBooth job '%d' is already used")
	errFusionBrainJobIsAlreadyUsed = errors.New("FusionBrain job '%d' is already used")
	chatIDIsNotExistErr            = errors.New("client with current chatID is not exist")
	chatIDAlreadyExistErr          = errors.New("client with current chatID already exist")
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

func ErrorFusionBrainJobIsNotExist(id int) error {
	return fmt.Errorf(errFusionBrainJobIsNotExist.Error(), id)
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

func ErrorFusionBrainJobIsAlreadyUsed(id int) error {
	return fmt.Errorf(errFusionBrainJobIsAlreadyUsed.Error(), id)
}

type clientState struct {
	command        string
	username       string
	openAICancels  map[int]context.CancelFunc
	chatGPTCancels map[int]context.CancelFunc
	dbCancels      map[int]context.CancelFunc
	fbCancels      map[int]context.CancelFunc
}

func NewTClient(username string) *clientState {
	return &clientState{
		command:        commandStart,
		username:       username,
		openAICancels:  make(map[int]context.CancelFunc),
		chatGPTCancels: make(map[int]context.CancelFunc),
		dbCancels:      make(map[int]context.CancelFunc),
		fbCancels:      make(map[int]context.CancelFunc),
	}
}

func (c *clientState) LenOpenAIJobs() int {
	return len(c.openAICancels)
}

func (c *clientState) LenChatGPTJobs() int {
	return len(c.chatGPTCancels)
}

func (c *clientState) LenDreamBoothJobs() int {
	return len(c.dbCancels)
}

func (c *clientState) LenFusionBrainJobs() int {
	return len(c.dbCancels)
}

func (c *clientState) SetUsername(username string) {
	c.username = username
}

func (c *clientState) Username() string {
	return c.username
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
	jobIDs := make([]int, 0, len(c.dbCancels))
	for id := range c.dbCancels {
		jobIDs = append(jobIDs, id)
	}
	return jobIDs
}

func (c *clientState) FusionBrainJobs() []int {
	jobIDs := make([]int, 0, len(c.fbCancels))
	for id := range c.fbCancels {
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
	if _, ok := c.dbCancels[id]; ok {
		return ErrorDreamBoothJobIsAlreadyUsed(id)
	}
	c.dbCancels[id] = cancel
	return nil
}

func (c *clientState) SetCancelFusionBrainJob(cancel context.CancelFunc, id int) error {
	if _, ok := c.fbCancels[id]; ok {
		return ErrorFusionBrainJobIsAlreadyUsed(id)
	}
	c.fbCancels[id] = cancel
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
	cancel, ok := c.dbCancels[id]
	if !ok {
		return ErrorDreamBoothJobIsNotExist(id)
	}
	cancel()
	delete(c.dbCancels, id)
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

func (c *clientState) CancelFusionBrainJob(id int) error {
	cancel, ok := c.fbCancels[id]
	if !ok {
		return ErrorFusionBrainJobIsNotExist(id)
	}
	cancel()
	delete(c.fbCancels, id)
	return nil
}

func (c *clientState) CancelChatGPTJobs() {
	for _, cancel := range c.chatGPTCancels {
		cancel()
	}
	c.chatGPTCancels = make(map[int]context.CancelFunc)
}

func (c *clientState) CancelDreamBoothJobs() {
	for _, cancel := range c.dbCancels {
		cancel()
	}
	c.dbCancels = make(map[int]context.CancelFunc)
}

func (c *clientState) CancelOpenAIJobs() {
	for _, cancel := range c.openAICancels {
		cancel()
	}
	c.openAICancels = make(map[int]context.CancelFunc)
}

func (c *clientState) CancelFusionBrainJobs() {
	for _, cancel := range c.fbCancels {
		cancel()
	}
	c.fbCancels = make(map[int]context.CancelFunc)
}

type clientStateByChatID struct {
	value map[int64]*clientState
	mutex sync.RWMutex
}

func (c *clientStateByChatID) AddClient(chatID int64, username string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	_, ok := c.value[chatID]
	if ok {
		return chatIDAlreadyExistErr
	}
	c.value[chatID] = NewTClient(username)
	return nil
}

func (c *clientStateByChatID) DeleteClient(chatID int64) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	tc, ok := c.value[chatID]
	if !ok || tc == nil {
		return chatIDIsNotExistErr
	}
	delete(c.value, chatID)
	return nil
}

func (c *clientStateByChatID) UpdateClientCommand(chatID int64, command string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	tc, ok := c.value[chatID]
	if !ok || tc == nil {
		return chatIDIsNotExistErr
	}
	tc.SetCommand(command)
	return nil
}

func (c *clientStateByChatID) ClientCommand(chatID int64) (string, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	tc, ok := c.value[chatID]
	if !ok || tc == nil {
		return "", chatIDIsNotExistErr
	}
	return tc.Command(), nil
}

func (c *clientStateByChatID) ClientChatGPTJobs(chatID int64) ([]int, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	tc, ok := c.value[chatID]
	if !ok || tc == nil {
		return nil, chatIDIsNotExistErr
	}
	return tc.ChatGPTJobs(), nil
}

func (c *clientStateByChatID) ClientDreamBoothJobs(chatID int64) ([]int, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	tc, ok := c.value[chatID]
	if !ok || tc == nil {
		return nil, chatIDIsNotExistErr
	}
	return tc.DreamBoothJobs(), nil
}

func (c *clientStateByChatID) ClientOpenAIJobs(chatID int64) ([]int, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	tc, ok := c.value[chatID]
	if !ok || tc == nil {
		return nil, chatIDIsNotExistErr
	}
	return tc.OpenAIJobs(), nil
}

func (c *clientStateByChatID) ClientFusionBrainJobs(chatID int64) ([]int, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	tc, ok := c.value[chatID]
	if !ok || tc == nil {
		return nil, chatIDIsNotExistErr
	}
	return tc.FusionBrainJobs(), nil
}

func (c *clientStateByChatID) ClientLenChatGPTJobs(chatID int64) (int, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	tc, ok := c.value[chatID]
	if !ok || tc == nil {
		return -1, chatIDIsNotExistErr
	}
	return tc.LenChatGPTJobs(), nil
}

func (c *clientStateByChatID) ClientLenDreamBoothJobs(chatID int64) (int, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	tc, ok := c.value[chatID]
	if !ok || tc == nil {
		return -1, chatIDIsNotExistErr
	}
	return tc.LenDreamBoothJobs(), nil
}

func (c *clientStateByChatID) ClientLenOpenAIJobs(chatID int64) (int, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	tc, ok := c.value[chatID]
	if !ok || tc == nil {
		return -1, chatIDIsNotExistErr
	}
	return tc.LenOpenAIJobs(), nil
}

func (c *clientStateByChatID) ClientLenFusionBrainJobs(chatID int64) (int, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	tc, ok := c.value[chatID]
	if !ok || tc == nil {
		return -1, chatIDIsNotExistErr
	}
	return tc.LenFusionBrainJobs(), nil
}

func (c *clientStateByChatID) ClientAddChatGPTJob(cancel context.CancelFunc, jobID int, chatID int64) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	tc, ok := c.value[chatID]
	if !ok || tc == nil {
		return chatIDIsNotExistErr
	}
	return tc.SetCancelChatGPTJob(cancel, jobID)
}

func (c *clientStateByChatID) ClientAddDreamBoothJob(cancel context.CancelFunc, jobID int, chatID int64) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	tc, ok := c.value[chatID]
	if !ok || tc == nil {
		return chatIDIsNotExistErr
	}
	return tc.SetCancelDreamBoothJob(cancel, jobID)
}

func (c *clientStateByChatID) ClientAddOpenAIJob(cancel context.CancelFunc, jobID int, chatID int64) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	tc, ok := c.value[chatID]
	if !ok || tc == nil {
		return chatIDIsNotExistErr
	}
	return tc.SetCancelOpenAIJob(cancel, jobID)
}

func (c *clientStateByChatID) ClientAddFusionBrainJob(cancel context.CancelFunc, jobID int, chatID int64) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	tc, ok := c.value[chatID]
	if !ok || tc == nil {
		return chatIDIsNotExistErr
	}
	return tc.SetCancelFusionBrainJob(cancel, jobID)
}

func (c *clientStateByChatID) ClientCancelChatGPTJob(jobID int, chatID int64) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	tc, ok := c.value[chatID]
	if !ok || tc == nil {
		return chatIDIsNotExistErr
	}
	return tc.CancelChatGPTJob(jobID)
}

func (c *clientStateByChatID) ClientCancelDreamBoothJob(jobID int, chatID int64) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	tc, ok := c.value[chatID]
	if !ok || tc == nil {
		return chatIDIsNotExistErr
	}
	return tc.CancelDreamBoothJob(jobID)
}

func (c *clientStateByChatID) ClientCancelOpenAIJob(jobID int, chatID int64) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	tc, ok := c.value[chatID]
	if !ok || tc == nil {
		return chatIDIsNotExistErr
	}
	return tc.CancelOpenAIJob(jobID)
}

func (c *clientStateByChatID) ClientCancelFusionBrainJob(jobID int, chatID int64) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	tc, ok := c.value[chatID]
	if !ok || tc == nil {
		return chatIDIsNotExistErr
	}
	return tc.CancelFusionBrainJob(jobID)
}

func (c *clientStateByChatID) ClientCancelJobs(chatID int64) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	tc, ok := c.value[chatID]
	if !ok || tc == nil {
		return chatIDIsNotExistErr
	}
	tc.CancelChatGPTJobs()
	tc.CancelDreamBoothJobs()
	tc.CancelFusionBrainJobs()
	return nil
}

func (c *clientStateByChatID) ClientUsername(chatID int64) (string, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	tc, ok := c.value[chatID]
	if !ok || tc == nil {
		return "", chatIDIsNotExistErr
	}
	return tc.Username(), nil
}

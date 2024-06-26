package tbotopenai

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
)

const fileNameStats = "Статистика_запросов.csv"

var csvHeader = []string{"Время", "Имя пользователя", "AI", "Запрос", "Ответ"}

type statRow struct {
	ts       string
	username string
	ai       string
	request  string
	response string
}

type statRows []statRow

func (s *statRows) Unmarshal() [][]string {
	rows := make([][]string, 0, len(*s))
	for idx := range *s {
		row := make([]string, 0, 5)
		row = append(row, (*s)[idx].ts)
		row = append(row, (*s)[idx].username)
		row = append(row, (*s)[idx].ai)
		row = append(row, (*s)[idx].request)
		row = append(row, (*s)[idx].response)
		rows = append(rows, row)
	}
	return rows
}

func (s *statRows) Flush() {
	*s = []statRow{}
}

type Stats struct {
	ticker   *time.Ticker
	filepath string
	rows     statRows
	buf      []byte
	quitChan chan struct{}
	log      *zap.Logger
}

func NewStats(log *zap.Logger, interval time.Duration, filepath string) *Stats {
	return &Stats{
		ticker:   time.NewTicker(interval),
		filepath: filepath,
		quitChan: make(chan struct{}, 1),
		log:      log,
	}
}

func (s *Stats) Run(wg *sync.WaitGroup) error {
	if err := os.MkdirAll(filepath.Dir(s.filepath), os.ModePerm); err != nil {
		return err
	}
	if _, err := os.Stat(s.filepath); os.IsNotExist(err) {
		var file *os.File
		file, err = os.OpenFile(s.filepath, os.O_WRONLY|os.O_CREATE, os.ModePerm)
		if err != nil {
			return err
		}
		w := csv.NewWriter(file)
		if err = w.Write(csvHeader); err != nil {
			return err
		}
		w.Flush()
		if err = file.Close(); err != nil {
			return err
		}
	}
	wg.Add(1)
	go s.initStatRowsWorker(wg)
	return nil
}

func (s *Stats) Stop() {
	close(s.quitChan)
}

func (s *Stats) initStatRowsWorker(wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			s.log.Error("Recovered panic err:", zap.Any("panic", r))
		}
	}()
	defer wg.Done()
	for {
		select {
		case _, ok := <-s.ticker.C:
			if !ok {
				return
			}
			if err := s.processStatRows(); err != nil {
				s.log.Error("process stat's rows err:", zap.Error(err))
			}
		case <-s.quitChan:
			s.ticker.Stop()
			return
		}
	}
}

func (s *Stats) processStatRows() error {
	file, err := os.OpenFile(s.filepath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}
	w := csv.NewWriter(file)
	for _, row := range s.rows.Unmarshal() {
		if err = w.Write(row); err != nil {
			return err
		}
	}
	s.rows.Flush()
	w.Flush()
	if err = file.Close(); err != nil {
		return err
	}
	if s.buf, err = os.ReadFile(s.filepath); err != nil {
		return err
	}
	return nil
}

func (s *Stats) Write(row statRow) {
	s.rows = append(s.rows, row)
}

func (s *Stats) Bytes() []byte {
	return s.buf
}

package main

import (
	"container/list"
	"errors"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/Songmu/strrand"
)

type GomibakoKey string

type GomibakoRequest struct {
	timestamp     time.Time
	method        string
	url           *url.URL
	headers       http.Header
	body          []byte
	contentLength int64
}

type Gomibako struct {
	key      GomibakoKey
	requests *list.List
}

type GomibakoRepository struct {
	gomibakos map[GomibakoKey]*Gomibako
	mutex     sync.RWMutex
}

func NewGomibakoRepository() *GomibakoRepository {
	gr := GomibakoRepository{
		gomibakos: make(map[GomibakoKey]*Gomibako),
		mutex:     sync.RWMutex{},
	}
	return &gr
}

func (gr *GomibakoRepository) AddGomibako() (*Gomibako, error) {
	gr.mutex.Lock()
	defer gr.mutex.Unlock()

	str, err := strrand.RandomString("[a-z0-9]{10}")
	if err != nil {
		return nil, err
	}
	newKey := GomibakoKey(str)

	gr.gomibakos[newKey] = &Gomibako{
		key:      newKey,
		requests: list.New(),
	}

	return gr.gomibakos[newKey], nil
}

func (gr *GomibakoRepository) AddRequest(key GomibakoKey, greq *GomibakoRequest) error {
	gr.mutex.Lock()
	defer gr.mutex.Unlock()

	g, ok := gr.gomibakos[key]
	if !ok {
		return errors.New("no gomibako found")
	}
	g.requests.PushBack(greq)
	if g.requests.Len() > 10 {
		g.requests.Remove(g.requests.Front())
	}
	return nil
}

func (gr *GomibakoRepository) Get(key GomibakoKey) (*Gomibako, error) {
	gr.mutex.RLock()
	defer gr.mutex.RUnlock()

	g, ok := gr.gomibakos[key]
	if !ok {
		return nil, errors.New("no gomibako found")
	}
	return g, nil
}

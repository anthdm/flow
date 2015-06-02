package proxy

import (
	"log"
	"sync"
)

type Muxer interface {
	Register(route string, name string)
	GetName(name string) (string, bool)
}

type mux struct {
	mu       sync.RWMutex
	routeMap map[string]string
}

func NewMux() *mux {
	return &mux{
		routeMap: map[string]string{},
	}
}

func (m *mux) Register(route, svcName string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.routeMap[route] = svcName
	log.Printf("mapped route %s => %s", route, svcName)
}

func (m *mux) GetName(svcName string) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	name, ok := m.routeMap[svcName]
	return name, ok
}

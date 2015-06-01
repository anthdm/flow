package proxy

import (
	"errors"
	"log"
	"sync"
)

var (
	errMissingService   = errors.New("missing service")
	errMissingEndpoints = errors.New("missing endpoints")
)

// LoadBalancer is anything that can return a endpoint
type LoadBalancer interface {
	AddService(svcName string)
	UpdateState(svcName string, endpoints []string)
	NextEndpoint(svcName string) (string, error)
}

// serviceBalancer directs traffic between nodes in the service cluster
type serviceBalancer struct {
	lock     sync.RWMutex
	services map[string]*balancerState
}

// balancerState keeps track of service endpoints and their current state
type balancerState struct {
	endpoints []string
	index     int
}

func NewServiceBalancer() *serviceBalancer {
	return &serviceBalancer{
		services: map[string]*balancerState{},
	}
}

func (sb *serviceBalancer) NextEndpoint(svcName string) (string, error) {
	sb.lock.Lock()
	defer sb.lock.Unlock()

	state, exists := sb.services[svcName]
	if !exists {
		log.Println("%v does not exist in %v", state, sb.services)
		return "", errMissingService
	}
	if len(state.endpoints) == 0 {
		return "", errMissingEndpoints
	}
	endpoint := state.endpoints[state.index]
	log.Printf("serving %s for service %s", endpoint, svcName)
	state.index = (state.index + 1) % len(state.endpoints)
	return endpoint, nil
}

func (sb *serviceBalancer) AddService(svcName string) {
	sb.lock.Lock()
	defer sb.lock.Unlock()

	_, exists := sb.services[svcName]
	if exists {
		panic("service allready registered")
	}
	log.Printf("registered service %s to the loadbalancer", svcName)
	sb.services[svcName] = &balancerState{}
}

// TODO: loop state endpoints and only remove or add those missing or present
// asumes the lock is allready held
func (sb *serviceBalancer) UpdateState(svcName string, endpoints []string) {
	sb.lock.Lock()
	defer sb.lock.Unlock()

	state, exists := sb.services[svcName]
	if !exists {
		panic(errMissingService)
	}
	state.endpoints = endpoints
}

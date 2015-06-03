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
	AddService(service ServicePortName)
	UpdateState(service ServicePortName, endpoints []string)
	NextEndpoint(service ServicePortName) (string, error)
}

// ServicePortName is an unique identifier for a registered service
type ServicePortName struct {
	// A service is assumed to have its proxy port on the same machine flow is
	// running
	Port string
	Name string
}

// serviceBalancer directs traffic between nodes from the same service cluster
type serviceBalancer struct {
	lock     sync.RWMutex
	services map[ServicePortName]*balancerState
}

// balancerState keeps track of service endpoints and their index
type balancerState struct {
	endpoints []string
	index     int
}

func NewServiceBalancer() *serviceBalancer {
	return &serviceBalancer{
		services: map[ServicePortName]*balancerState{},
	}
}

func (sb *serviceBalancer) NextEndpoint(service ServicePortName) (string, error) {
	sb.lock.Lock()
	defer sb.lock.Unlock()

	state, exists := sb.services[service]
	if !exists {
		log.Println("%v does not exist in %v", state, sb.services)
		return "", errMissingService
	}
	if len(state.endpoints) == 0 {
		return "", errMissingEndpoints
	}
	endpoint := state.endpoints[state.index]
	log.Printf("serving %s for service %s", endpoint, service)
	state.index = (state.index + 1) % len(state.endpoints)
	return endpoint, nil
}

func (sb *serviceBalancer) AddService(service ServicePortName) {
	sb.lock.Lock()
	defer sb.lock.Unlock()

	_, exists := sb.services[service]
	if exists {
		panic("service allready registered")
	}
	log.Printf("registered %s to the loadbalancer", service)
	sb.services[service] = &balancerState{}
}

// TODO: loop state endpoints and only remove or add those missing or present
// asumes the lock is allready held
func (sb *serviceBalancer) UpdateState(service ServicePortName, endpoints []string) {
	sb.lock.Lock()
	defer sb.lock.Unlock()

	state, exists := sb.services[service]
	if !exists {
		panic(errMissingService)
	}
	state.endpoints = endpoints
}

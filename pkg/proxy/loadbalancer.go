package proxy

import (
	"errors"
	"log"
	"sync"
)

var errMissingService = errors.New("missing service")

// LoadBalancer is anything that can return a endpoint
type LoadBalancer interface {
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

func (sb *serviceBalancer) NextEndpoint(svcName string) (string, error) {
	sb.lock.Lock()
	defer sb.lock.Unlock()

	state, exists := sb.services[svcName]
	if !exists {
		return "", errMissingService
	}
	endpoint := state.endpoints[state.index]
	log.Printf("serving %s for service %s", endpoint, svcName)
	state.index = (state.index + 1) % len(state.endpoints)
	return endpoint, nil
}

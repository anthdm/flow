package proxy

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"sync"

	"github.com/twanies/flow/api"
)

var (
	errMissingService   = errors.New("missing service")
	errMissingEndpoints = errors.New("missing endpoints")
)

// LoadBalancer is an interface for directing traffic between service endpoints
type LoadBalancer interface {
	AddService(service ServicePortName)
	NextEndpoint(service ServicePortName) (string, error)
}

// ServicePortName is an unique identifier for a registered service
type ServicePortName struct {
	// A service is assumed to have its proxy port on the same machine flow is
	// running
	Name string
	Port string
}

func (s ServicePortName) String() string {
	return fmt.Sprintf("%s:%s", s.Name, s.Port)
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
	sb.addServiceInternal(service)
}

// addServiceInternal used for adding a new service when the lock is allready held
// Addservice will cause a deadlock assumes the lock is allready held
func (sb *serviceBalancer) addServiceInternal(service ServicePortName) *balancerState {
	if _, exists := sb.services[service]; !exists {
		sb.services[service] = &balancerState{}
	}
	log.Printf("registered %s to the loadbalancer", service)
	return sb.services[service]
}

// Update wil compare the new endpointSet with the existing state.
func (sb *serviceBalancer) Update(endpoints []api.Endpoints) {
	registeredEndpoints := make(map[ServicePortName]bool)
	sb.lock.Lock()
	defer sb.lock.Unlock()

	for i := range endpoints {
		svcEndpoints := &endpoints[i]
		serviceName := ServicePortName{svcEndpoints.Name, ""}
		state, exists := sb.services[serviceName]
		curEndpoints := []string{}
		if state != nil {
			curEndpoints = state.endpoints
		}
		newEndpoints := endpointsToSlice(svcEndpoints)
		if !exists || !equalSlices(curEndpoints, newEndpoints) {
			state := sb.addServiceInternal(serviceName)
			state.endpoints = endpointsToSlice(svcEndpoints)
			state.index = 0
		}
		registeredEndpoints[serviceName] = true
	}

	for k := range sb.services {
		if _, ok := registeredEndpoints[k]; !ok {
			delete(sb.services, k)
		}
	}
}

func endpointsToSlice(endpoints *api.Endpoints) []string {
	var out []string
	for _, endpoint := range endpoints.Subset {
		out = append(out, fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port))
	}
	return out
}

func equalSlices(src, dst []string) bool {
	if len(src) != len(dst) {
		return false
	}
	if !reflect.DeepEqual(src, dst) {
		return false
	}
	return true
}

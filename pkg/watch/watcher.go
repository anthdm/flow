package watch

import (
	"github.com/twanies/flow/api"
	"github.com/twanies/flow/pkg/registry"
)

// ServiceUpdateHandler handles updating the current state to the desired state
// of services
type ServiceUpdateHandler interface {
	Update(services []api.Service)
}

type ServiceUpdate struct {
	services []api.Service
}

// ServiceWatcher watches for changes in the registry, it invokes the update handler
// Update method when a change is detected
type ServiceWatcher struct {
	store   registry.Register
	handler ServiceUpdateHandler
}

func NewServiceWatcher() *ServiceWatcher {
	store := registry.NewRegistry()
	return &ServiceWatcher{
		store: store,
	}
}

func (sw *ServiceWatcher) RegisterHandler(handler ServiceUpdateHandler) {
	sw.handler = handler
	go sw.WatchForUpdates()
}

func (sw *ServiceWatcher) WatchForUpdates() {
	serviceUpdate := make(chan []api.Service)
	go sw.store.WatchServices(serviceUpdate)
	for true {
		services := <-serviceUpdate
		sw.handler.Update(services)
	}
}

type EndpointUpdateHandler interface {
	Update(endpoints []api.Endpoints)
}

type EndpointWatcher struct {
	store   registry.Register
	handler EndpointUpdateHandler
}

func NewEndpointWatcher() *EndpointWatcher {
	store := registry.NewRegistry()
	return &EndpointWatcher{store: store}
}

func (ew *EndpointWatcher) RegisterHandler(handler EndpointUpdateHandler) {
	ew.handler = handler
	go ew.WatchForUpdates()
}

func (ew *EndpointWatcher) WatchForUpdates() {

}

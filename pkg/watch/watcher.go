package watch

import (
	"github.com/twanies/flow/api"
	"github.com/twanies/flow/pkg/registry"
)

type ServiceUpdateHandler interface {
	Update(services []api.Service)
}

type ServiceUpdate struct {
	services []api.Service
}

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

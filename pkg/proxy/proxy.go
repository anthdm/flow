package proxy

// import (
// 	"errors"
// 	"log"
// 	"sync"

// 	"github.com/twanies/flow/api"
// 	"github.com/twanies/flow/pkg/registry"
// )

// var errMissingRoute = errors.New("missing route")

// type serviceInfo struct {
// 	name     string
// 	frontend api.FrontendMeta
// 	protocol string
// 	nodes    api.NodeList
// }

// type Proxy struct {
// 	balancer   LoadBalancer
// 	mux        Muxer
// 	registry   registry.Register
// 	updateChan chan api.ServiceList

// 	lock       sync.RWMutex // protects following
// 	serviceMap map[string]*serviceInfo
// }

// func New() *Proxy {
// 	return &Proxy{
// 		serviceMap: map[string]*serviceInfo{},
// 		updateChan: make(chan api.ServiceList),
// 		balancer:   NewServiceBalancer(),
// 		mux:        NewMux(),
// 	}
// }

// // Discover starts to subcribe on the registry and starts a goroutine to watch
// // for changes on services
// func (p *Proxy) Discover(registry registry.Register) {
// 	p.registry = registry
// 	go p.registry.Subscribe(p.updateChan)
// 	go p.discoveryLoop()
// }

// // discoveryLoop watches for updates in the registry
// func (p *Proxy) discoveryLoop() {
// 	for {
// 		select {
// 		case services := <-p.updateChan:
// 			log.Printf("discovered service changes, updating..")
// 			p.Update(services)
// 		}
// 	}
// }

// func (p *Proxy) getServiceInfo(svcName string) (*serviceInfo, bool) {
// 	p.lock.Lock()
// 	defer p.lock.Unlock()
// 	info, ok := p.serviceMap[svcName]
// 	return info, ok
// }

// func (p *Proxy) setServiceInfo(svcInfo *serviceInfo) {
// 	p.lock.Lock()
// 	defer p.lock.Unlock()
// 	p.serviceMap[svcInfo.name] = svcInfo
// }

// // Delete a service completly from the proxy
// // asumes the lock is already held
// func (p *Proxy) deleteService(svcName string, info *serviceInfo) {
// 	delete(p.serviceMap, svcName)
// 	p.mux.Unregister(info.frontend.Route)
// }

// func (p *Proxy) Update(services api.ServiceList) {
// 	activeServices := make(map[string]bool)
// 	for i := range services {
// 		service := &services[i]
// 		serviceInfo := &serviceInfo{
// 			name:     service.Name,
// 			frontend: service.Frontend,
// 			protocol: service.Protocol,
// 			nodes:    service.Nodes,
// 		}
// 		endpoints := make([]string, len(service.Nodes))
// 		for i, n := range service.Nodes {
// 			endpoints[i] = n.String()
// 		}
// 		activeServices[service.Name] = true

// 		info, exists := p.getServiceInfo(service.Name)
// 		if exists && sameInfo(info, service) {
// 			// no changes
// 			continue
// 		}
// 		if exists {
// 			// update the serviceMap and only update the balancerState
// 			log.Printf("receiving update for service %s", service.Name)
// 			p.setServiceInfo(serviceInfo)
// 			p.balancer.UpdateState(serviceInfo.name, endpoints)
// 			continue
// 		}
// 		// A new service is registering
// 		log.Printf("registering %s as a new service", service.Name)
// 		p.mux.Register(service.Frontend.Route, service.Name)
// 		p.serviceMap[service.Name] = serviceInfo
// 		p.balancer.AddService(serviceInfo.name)
// 		p.balancer.UpdateState(serviceInfo.name, endpoints)
// 	}

// 	// removing all services not in the servicelist
// 	p.lock.Lock()
// 	defer p.lock.Unlock()
// 	for name, info := range p.serviceMap {
// 		if !activeServices[name] {
// 			log.Printf("removing %s from the proxy", name)
// 			p.deleteService(name, info)
// 		}
// 	}
// }

// func sameInfo(info *serviceInfo, service *api.Service) bool {
// 	if info.frontend.Route != service.Frontend.Route {
// 		return false
// 	}
// 	if info.frontend.Scheme != service.Frontend.Scheme {
// 		return false
// 	}
// 	if info.frontend.TargetPath != service.Frontend.TargetPath {
// 		return false
// 	}
// 	if info.protocol != service.Protocol {
// 		return false
// 	}
// 	if !equalBalancerState() {
// 		return false
// 	}
// 	return true
// }

// func equalBalancerState() bool {
// 	return true
// }

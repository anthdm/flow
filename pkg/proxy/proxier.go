package proxy

import (
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/twanies/flow/api"
)

type serviceInfo struct {
	protocol string
	socket   ProxySocket

	// proxyPort is the port assigned by flow where the service proxy wil listen on
	proxyPort int
}

// Proxier proxies incomming traffic between its endpoints
type Proxier struct {
	loadBalancer LoadBalancer

	// number of accepted connections in the proxyLoop. Atomicly updated
	numLoops int32

	mu         sync.RWMutex // protects following
	serviceMap map[ServicePortName]*serviceInfo
	proxyPorts *PortAllocator
}

func NewProxier(loadBalancer LoadBalancer) *Proxier {
	proxyPorts := NewPortAllocator(2000, 3000)
	return &Proxier{
		loadBalancer: loadBalancer,
		serviceMap:   make(map[ServicePortName]*serviceInfo),
		proxyPorts:   proxyPorts,
	}
}

// Update will sync the endpoints to their new state
func (p *Proxier) Update(services []api.Service) {
	activeServices := make(map[ServicePortName]bool)
	for i := range services {
		service := &services[i]
		serviceName := ServicePortName{service.Name, ""}
		info, exists := p.getServiceInfo(serviceName)
		if exists && sameInfo(info, service) {
			// no updates
			continue
		}
		if exists {
			log.Printf("receiving updates for service %s", serviceName)
			panic("no implementation")

		}
		log.Printf("discovering %s as a new service", serviceName)
		port, err := p.proxyPorts.AssignNext()
		if err != nil {
			log.Printf("failed to assign new port for %s", serviceName)
			continue
		}
		info, err = p.addServiceToPort(serviceName, service.Protocol, port)
		if err != nil {
			log.Printf("failed to start proxy for %s: %v", serviceName, err)
			continue
		}
		log.Printf("service %s running on port %d", serviceName, port)
		p.loadBalancer.AddService(serviceName)
		activeServices[serviceName] = true
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	for service, info := range p.serviceMap {
		if !activeServices[service] {
			// Stop the service internal
			log.Printf("stopping service %s", service)
			delete(p.serviceMap, service)
			if err := info.socket.Close(); err != nil {
				log.Printf("failed to stop service %s", service)
			}
			p.proxyPorts.Release(info.proxyPort)
		}
	}
}

func sameInfo(info *serviceInfo, service *api.Service) bool {
	if info.protocol != service.Protocol || info.proxyPort != service.Port {
		return false
	}
	return true
}

func (p *Proxier) getServiceInfo(service ServicePortName) (*serviceInfo, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	info, ok := p.serviceMap[service]
	return info, ok
}

func (p *Proxier) setServiceInfo(service ServicePortName, info *serviceInfo) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.serviceMap[service] = info
}

func (p *Proxier) Discover() {
	for {
		select {
		default:
			time.Sleep(10 * time.Second)
			log.Println("looping")
		}
	}
}

func (p *Proxier) addServiceToPort(service ServicePortName, protocol string, proxyPort int) (*serviceInfo, error) {
	sock, err := newProxySocket(protocol, proxyPort)
	if err != nil {
		return nil, err
	}
	info := &serviceInfo{
		protocol:  protocol,
		proxyPort: proxyPort,
		socket:    sock,
	}
	p.setServiceInfo(service, info)

	go func(service ServicePortName, p *Proxier) {
		atomic.AddInt32(&p.numLoops, 1)
		sock.ProxyLoop(service, info, p)
		atomic.AddInt32(&p.numLoops, -1)
	}(service, p)
	return info, nil
}

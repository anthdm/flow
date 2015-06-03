package proxy

import (
	"log"
	"sync"
	"time"
)

type serviceInfo struct {
	protocol string
	port     int
}

// Proxier proxies incomming traffic between its endpoints
type Proxier struct {
	loadBalancer LoadBalancer

	mu         sync.RWMutex // protects following
	serviceMap map[ServicePortName]*serviceInfo
}

func NewProxier(loadBalancer LoadBalancer) *Proxier {
	return &Proxier{
		loadBalancer: loadBalancer,
		serviceMap:   make(map[ServicePortName]*serviceInfo),
	}
}

func (p *Proxier) getServiceInfo(name ServicePortName) (*serviceInfo, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	info, ok := p.serviceMap[name]
	return info, ok
}

func (p *Proxier) setServiceInfo(name ServicePortName, info *serviceInfo) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.serviceMap[name] = info
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

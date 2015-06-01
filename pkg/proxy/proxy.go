package proxy

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/twanies/flow/api"
)

type serviceInfo struct {
	name     string
	frontend api.FrontendMeta
	protocol string
	nodes    api.NodeList
}

type node struct {
	endpoint hostPortPair
}

type hostPortPair struct {
	host string
	port int
}

func (hpp hostPortPair) String() string {
	return fmt.Sprintf("%s:%d", hpp.host, hpp.port)
}

type Proxy struct {
	balancer LoadBalancer

	lock       sync.RWMutex // protects followin
	serviceMap map[string]*serviceInfo
}

func New() *Proxy {
	return &Proxy{
		serviceMap: map[string]*serviceInfo{},
		balancer:   NewServiceBalancer(),
	}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	svc, exists := p.serviceMap[r.RequestURI]
	if !exists {
		http.Error(w, errMissingService.Error(), http.StatusBadRequest)
		return
	}
	endpoint, err := p.balancer.NextEndpoint(svc.name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	path := svc.frontend.TargetPath
	if path == "" {
		path = "/"
	}

	r.URL.Scheme = svc.frontend.Scheme
	r.URL.Host = endpoint
	r.URL.Path = path

	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		log.Printf("WARNING %v", err.Error())
		p.ServeHTTP(w, r)
		return
	}
	io.Copy(w, resp.Body)
}

func (p *Proxy) Update(services []api.Service) {
	p.lock.Lock()
	defer p.lock.Unlock()

	for _, svc := range services {
		_, exists := p.serviceMap[svc.Name]
		if exists {
			// TODO update
			log.Printf("receiving update for service %s", svc.Name)
			return
		}
		log.Printf("receiving new service %s", svc.Name)
		endpoints := make([]string, len(svc.Nodes))
		for i, n := range svc.Nodes {
			endpoints[i] = n.Endpoint.String()
		}
		serviceInfo := &serviceInfo{
			name:     svc.Name,
			frontend: svc.Frontend,
			protocol: svc.Protocol,
			nodes:    svc.Nodes,
		}

		p.serviceMap[svc.Frontend.Route] = serviceInfo
		// TODO only if its a new service
		p.balancer.AddService(serviceInfo.name)
		p.balancer.UpdateState(serviceInfo.name, endpoints)
	}
}

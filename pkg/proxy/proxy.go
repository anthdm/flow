package proxy

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/twanies/flow/api"
)

var errMissingRoute = errors.New("missing route")

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
	mux      Muxer

	lock       sync.RWMutex // protects following
	serviceMap map[string]*serviceInfo
}

func New() *Proxy {
	return &Proxy{
		serviceMap: map[string]*serviceInfo{},
		balancer:   NewServiceBalancer(),
		mux:        NewMux(),
	}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	svcName, ok := p.mux.GetName(r.RequestURI)
	if !ok {
		http.Error(w, errMissingRoute.Error(), http.StatusBadRequest)
		return
	}
	info, exists := p.getServiceInfo(svcName)
	if !exists {
		http.Error(w, errMissingService.Error(), http.StatusBadRequest)
		return
	}
	endpoint, err := p.balancer.NextEndpoint(info.name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	path := info.frontend.TargetPath
	if path == "" {
		path = "/"
	}

	r.URL.Scheme = info.frontend.Scheme
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

func (p *Proxy) getServiceInfo(svcName string) (*serviceInfo, bool) {
	p.lock.Lock()
	defer p.lock.Unlock()
	info, ok := p.serviceMap[svcName]
	return info, ok
}

func (p *Proxy) Update(services []api.Service) {
	for _, svc := range services {
		serviceInfo := &serviceInfo{
			name:     svc.Name,
			frontend: svc.Frontend,
			protocol: svc.Protocol,
			nodes:    svc.Nodes,
		}
		endpoints := make([]string, len(svc.Nodes))
		for i, n := range svc.Nodes {
			endpoints[i] = n.String()
		}

		_, exists := p.getServiceInfo(svc.Name)
		// TODO: compare these
		// update the serviceMap and only update the balancerState
		if exists {
			log.Printf("receiving update for service %s", svc.Name)
			p.serviceMap[svc.Name] = serviceInfo
			p.balancer.UpdateState(serviceInfo.name, endpoints)
			continue
		}
		log.Printf("registering %s as a new service", svc.Name)
		p.mux.Register(svc.Frontend.Route, svc.Name)
		p.serviceMap[svc.Name] = serviceInfo
		p.balancer.AddService(serviceInfo.name)
		p.balancer.UpdateState(serviceInfo.name, endpoints)
	}
}

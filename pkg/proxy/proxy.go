package proxy

import (
	"fmt"
)

type service struct {
	frontend frontendMetaData
	protocol string
	nodes    []*node
}

type frontendMetaData struct {
	route      string
	targetPath string
	scheme     string
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
}

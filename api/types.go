package api

import "fmt"

type Service struct {
	Name     string
	Frontend FrontendMeta
	Protocol string
	Nodes    NodeList
}

type Node struct {
	Endpoint HostPortPair
}

type NodeList []Node

type HostPortPair struct {
	Host string
	Port int
}

func (hpp HostPortPair) String() string {
	return fmt.Sprintf("%s:%d", hpp.Host, hpp.Port)
}

type FrontendMeta struct {
	Scheme     string
	TargetPath string
	Route      string
}

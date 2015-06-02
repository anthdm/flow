package api

import "fmt"

type Version struct {
	Version    string
	ApiVersion string
	GitCommit  string
}

type Service struct {
	Name     string       `json:"name"`
	Frontend FrontendMeta `json:"frontend"`
	Protocol string       `json:"protocol"`
	Nodes    NodeList     `json:"nodes"`
}

type ServiceList []Service

type Node struct {
	Host string
	Port int
}

func (n *Node) String() string {
	return fmt.Sprintf("%s:%d", n.Host, n.Port)
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
	Scheme     string `json:"scheme"`
	TargetPath string `json:"targetPath"`
	Route      string `json:"route"`
}

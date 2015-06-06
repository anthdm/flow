package api

type Version struct {
	Version    string
	ApiVersion string
	GitCommit  string
}

type Service struct {
	// Name is the human friendly DNS name of the service. Flow will map the name
	// to a valid proxy address.
	Name string `json:"name"`

	// Protocol is the IP protocol of the port. UDP" and "TCP"
	Protocol string

	// Port is the internal allocated port
	Port     int          `json:"port"`
	Frontend FrontendSpec `json:"frontend"`

	// TODO: implement multiple proxys for services that are not needed to be
	// loadbalanced like a mongo or redis server. These services need specific
	// ports to be claimed and assigned
	Ports []ServicePort `json:"ports"`
}

type ServicePort struct {
	// Port needed to be exposed for the service
	Port int

	// Protocol is the IP protocol of the port. UDP" and "TCP"
	Protocol string
}

// FrontendSpec lets us map HTTP requests to a specific service.
//
// EX.
// You can map "/v1/api" to a specific service. Flow wil handle the loadbalancing
// between the service endpoints.
type FrontendSpec struct {
	// HTTP scheme of the request that needs to be proxied. "HTTP" and "HTTPS"
	Scheme string `json:"scheme"`

	// TargetPath lets you specify a request URI. Default the TargetPath is "/"
	// the route "/v1/api" wil map to the root of your service endpoints.
	// TargetPath wil join the root "/"
	TargetPath string `json:"targetPath"`

	// Route is the request URI that wil map to the assigned servicePort
	Route string `json:"route"`
}

type Endpoint struct {
	Host string
	Port int
}

// EndpointSet is a set of endpoints implemented by a specific service
type EndpointSet struct {
	Name      string
	Endpoints []Endpoint
}

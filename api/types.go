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

	// ports to be claimed and assigned
	Ports []ServicePort `json:"ports"`
}

type ServicePort struct {
	// name of the port linked with the service
	Name string

	// Port needed to be exposed for the service
	Port int

	// TargetPort is the port exposed by the actual "container or process"
	TargetPort int

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

type EndpointPort struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

// Endpoints for a service
type Endpoints struct {
	Name      string         `json:"name"`
	Addresses []string       `json:"addresses"`
	Ports     []EndpointPort `json:"ports"`
}

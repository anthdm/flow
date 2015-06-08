package proxy

import (
	"testing"

	"github.com/twanies/flow/api"
)

func TestNewService(t *testing.T) {
	serviceName := ServicePortName{"fooo", "bar"}
	balancer := NewServiceBalancer()
	balancer.AddService(serviceName)
	_, ok := balancer.services[serviceName]
	if !ok {
		t.Errorf("expected balancer to have a state for %s", serviceName)
	}
}

func TestLoadBalancerFailsNoEndpoints(t *testing.T) {
	var endpoints []api.Endpoints
	loadBalancer := NewServiceBalancer()
	loadBalancer.Update(endpoints)
	service := ServicePortName{"foo", "bar"}
	endpoint, err := loadBalancer.NextEndpoint(service)
	if err == nil {
		t.Error("loadBalancer didnt fail with no endpoints")
	}
	if len(endpoint) != 0 {
		t.Error("got an endpoint")
	}
}

func TestExpectEndpoints(t *testing.T) {
	serviceName := ServicePortName{"foo", "a"}
	endpoints := api.Endpoints{
		Name:      "foo",
		Addresses: []string{"1.1"},
		Ports: []api.EndpointPort{
			api.EndpointPort{"a", 8080},
			api.EndpointPort{"a", 8081},
			api.EndpointPort{"a", 8082},
		},
	}

	balancer := NewServiceBalancer()
	balancer.AddService(serviceName)
	balancer.Update([]api.Endpoints{endpoints})
	expectEndpoint(t, serviceName, balancer, "1.1:8080")
	expectEndpoint(t, serviceName, balancer, "1.1:8081")
	expectEndpoint(t, serviceName, balancer, "1.1:8082")
	expectEndpoint(t, serviceName, balancer, "1.1:8080")
}

func TestExpectMultipleEndpointsMultiplePorts(t *testing.T) {
	serviceName1 := ServicePortName{"foo", "a"}
	serviceName2 := ServicePortName{"foo", "b"}
	endpoints := api.Endpoints{
		Name:      "foo",
		Addresses: []string{"1.1", "1.2"},
		Ports: []api.EndpointPort{
			api.EndpointPort{"a", 8080},
			api.EndpointPort{"b", 8081},
		},
	}
	balancer := NewServiceBalancer()
	balancer.Update([]api.Endpoints{endpoints})

	curEndpoints := balancer.services[serviceName1].endpoints
	expectEndpoint(t, serviceName1, balancer, curEndpoints[0])
	expectEndpoint(t, serviceName1, balancer, curEndpoints[1])
	expectEndpoint(t, serviceName1, balancer, curEndpoints[0])

	curEndpoints = balancer.services[serviceName2].endpoints
	expectEndpoint(t, serviceName2, balancer, curEndpoints[0])
	expectEndpoint(t, serviceName2, balancer, curEndpoints[1])
	expectEndpoint(t, serviceName2, balancer, curEndpoints[0])
}

func TestExpectMultipleEndpointsAndPortsWithUpdate(t *testing.T) {
	serviceName1 := ServicePortName{"foo", "a"}
	serviceName2 := ServicePortName{"foo", "b"}
	endpoints := api.Endpoints{
		Name:      "foo",
		Addresses: []string{"1.1", "1.2"},
		Ports: []api.EndpointPort{
			api.EndpointPort{"a", 8080},
			api.EndpointPort{"b", 8081},
			api.EndpointPort{"a", 8082},
			api.EndpointPort{"b", 8082},
		},
	}
	balancer := NewServiceBalancer()
	balancer.Update([]api.Endpoints{endpoints})

	curEndpoints := balancer.services[serviceName1].endpoints
	expectEndpoint(t, serviceName1, balancer, curEndpoints[0])
	expectEndpoint(t, serviceName1, balancer, curEndpoints[1])
	expectEndpoint(t, serviceName1, balancer, curEndpoints[2])
	expectEndpoint(t, serviceName1, balancer, curEndpoints[3])
	expectEndpoint(t, serviceName1, balancer, curEndpoints[0])

	curEndpoints = balancer.services[serviceName2].endpoints
	expectEndpoint(t, serviceName2, balancer, curEndpoints[0])
	expectEndpoint(t, serviceName2, balancer, curEndpoints[1])
	expectEndpoint(t, serviceName2, balancer, curEndpoints[2])
	expectEndpoint(t, serviceName2, balancer, curEndpoints[3])
	expectEndpoint(t, serviceName2, balancer, curEndpoints[0])

	endpoints = api.Endpoints{
		Name:      "foo",
		Addresses: []string{"1.3", "1.4"},
		Ports: []api.EndpointPort{
			api.EndpointPort{"a", 1},
			api.EndpointPort{"b", 2},
		},
	}

	balancer.Update([]api.Endpoints{endpoints})
	curEndpoints = balancer.services[serviceName1].endpoints
	expectEndpoint(t, serviceName1, balancer, curEndpoints[0])
	expectEndpoint(t, serviceName1, balancer, curEndpoints[1])
	expectEndpoint(t, serviceName1, balancer, curEndpoints[0])
}

func expectEndpoint(t *testing.T, service ServicePortName, balancer *serviceBalancer, expected string) {
	endpoint, err := balancer.NextEndpoint(service)
	if err != nil {
		t.Fatal(err)
	}
	if expected != endpoint {
		t.Fatalf("expected %s got %s", expected, endpoint)
	}
}

func TestNonEqualSlices(t *testing.T) {
	s1 := []string{"a", "b", "c"}
	s2 := []string{"a", "b"}
	if equalSlices(s1, s2) {
		t.Errorf("expected slices to be non equal")
	}
}

func TestEqualNonEqualSlicesSameLen(t *testing.T) {
	s1 := []string{"a", "b", "t"}
	s2 := []string{"a", "b", "c"}
	if equalSlices(s2, s1) {
		t.Errorf("expected slices to be non equal")
	}
}

func TestEqualSlices(t *testing.T) {
	s1 := []string{"a", "b", "c"}
	s2 := []string{"a", "b", "c"}
	if !equalSlices(s1, s2) {
		t.Errorf("expected slices to be equal")
	}
}

func TestEqualSlicesUnsorted(t *testing.T) {
	s1 := []string{"b", "a", "c"}
	s2 := []string{"a", "b", "c"}
	if !equalSlices(s1, s2) {
		t.Errorf("expected slices to be equal")
	}
}

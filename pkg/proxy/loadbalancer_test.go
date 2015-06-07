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
	serviceName := ServicePortName{"foo", ""}
	endpoints := api.Endpoints{
		Name: "foo",
		Subset: []api.Endpoint{
			api.Endpoint{"1.1", 3000},
			api.Endpoint{"1.1", 3001},
			api.Endpoint{"1.1", 3002},
		},
	}
	balancer := NewServiceBalancer()
	balancer.AddService(serviceName)
	balancer.Update([]api.Endpoints{endpoints})
	expectEndpoint(t, serviceName, balancer, "1.1:3000")
	expectEndpoint(t, serviceName, balancer, "1.1:3001")
	expectEndpoint(t, serviceName, balancer, "1.1:3002")
}

func TestExpectEndpointsWithNoService(t *testing.T) {
	endpoints := api.Endpoints{
		Name: "foo",
		Subset: []api.Endpoint{
			api.Endpoint{"1.1", 3000},
			api.Endpoint{"1.1", 3001},
			api.Endpoint{"1.1", 3002},
		},
	}
	loadBalancer := NewServiceBalancer()
	loadBalancer.Update([]api.Endpoints{endpoints})
	service := ServicePortName{"foo", ""}
	expectEndpoint(t, service, loadBalancer, "1.1:3000")
	expectEndpoint(t, service, loadBalancer, "1.1:3001")
	expectEndpoint(t, service, loadBalancer, "1.1:3002")
}

func TestUpdateDeleteEndpoints(t *testing.T) {
	service := ServicePortName{"foo", ""}
	lb := NewServiceBalancer()
	lb.AddService(service)
	endpoints := api.Endpoints{
		Name: "bar",
		Subset: []api.Endpoint{
			api.Endpoint{"1.1", 3000},
			api.Endpoint{"1.1", 3001},
			api.Endpoint{"1.1", 3002},
		},
	}
	lb.Update([]api.Endpoints{endpoints})
	if _, ok := lb.services[service]; ok {
		t.Fatal("expexted %s not to be present in the serviceMap")
	}
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

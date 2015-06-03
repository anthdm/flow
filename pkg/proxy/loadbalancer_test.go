package proxy

import (
	"testing"
)

func TestExpectEndpoints(t *testing.T) {
	endpoints := []string{"1:3000", "1:3001", "1:3002"}
	sb := &serviceBalancer{
		services: make(map[ServicePortName]*balancerState),
	}
	scv := ServicePortName{"chatservice", ":3000"}
	sb.services[scv] = &balancerState{endpoints: endpoints}
	expectEndpoint(t, scv, sb, "1:3000")
	expectEndpoint(t, scv, sb, "1:3001")
	expectEndpoint(t, scv, sb, "1:3002")
}

func TestAddService(t *testing.T) {
	sb := NewServiceBalancer()
	svc := ServicePortName{"foo", "bar"}
	sb.AddService(svc)
	_, exists := sb.services[svc]
	if !exists {
		t.Error("expected myservice to have a state")
	}
}

func TestUpdateNodes(t *testing.T) {
	sb := NewServiceBalancer()
	svc := ServicePortName{"foo", "bar"}
	sb.AddService(svc)
	sb.UpdateState(svc, []string{"1:3000", "1:3001"})
	expectEndpoint(t, svc, sb, "1:3000")
	expectEndpoint(t, svc, sb, "1:3001")
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

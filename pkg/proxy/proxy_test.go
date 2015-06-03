package proxy

import "net/http/httptest"

type tstServer struct{ *httptest.Server }

// func makeServices() api.ServiceList {
// 	services := make([]api.Service, 1)
// 	services[0] = api.Service{
// 		Name:     "foo",
// 		Protocol: "tcp",
// 		Frontend: api.FrontendMeta{
// 			Scheme:     "http",
// 			Route:      "/bar",
// 			TargetPath: "/",
// 		},
// 		Nodes: []api.Node{api.Node{Host: "0.0", Port: 3000}},
// 	}
// 	return services
// }

// func TestDeleteService(t *testing.T) {
// 	p := New()
// 	services := makeServices()
// 	p.Update(services)
// 	service, ok := p.serviceMap["foo"]
// 	if !ok {
// 		t.Fatal("expected proxy to have service (foo)")
// 	}
// 	p.deleteService(service.name, service)
// 	_, exists := p.serviceMap["foo"]
// 	if exists {
// 		t.Fatalf("expected service (foo) to be deleted got %+v", p.serviceMap)
// 	}
// }

// func TestGetServiceInfo(t *testing.T) {
// 	p := New()
// 	services := makeServices()
// 	p.Update(services)
// 	_, ok := p.getServiceInfo("foo")
// 	if !ok {
// 		t.Fatal("exptected serviceInfo to be present")
// 	}
// }

// func TestSameInfo(t *testing.T) {
// 	svcInfo := &serviceInfo{
// 		name: "tstservice",
// 		frontend: api.FrontendMeta{
// 			Scheme:     "http",
// 			Route:      "/hello",
// 			TargetPath: "/",
// 		},
// 		protocol: "tcp",
// 	}
// 	service := &api.Service{
// 		Name:     "tstservice",
// 		Protocol: "tcp",
// 		Frontend: api.FrontendMeta{
// 			Scheme:     "http",
// 			Route:      "/hello",
// 			TargetPath: "/",
// 		},
// 	}
// 	if !sameInfo(svcInfo, service) {
// 		t.Errorf("expected equal %+v, %+v", svcInfo, service)
// 	}
// 	service.Protocol = "udp"
// 	if sameInfo(svcInfo, service) {
// 		t.Errorf("expected not equal %+v, %+v", svcInfo, service)
// 	}
// }

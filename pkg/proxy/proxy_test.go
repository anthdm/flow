package proxy

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/twanies/flow/api"
)

type tstServer struct{ *httptest.Server }

func TestGetServiceInfo(t *testing.T) {
	p := New()
	services := make([]api.Service, 1)
	services[0] = api.Service{
		Name:     "myservice",
		Protocol: "tcp",
		Frontend: api.FrontendMeta{
			Scheme:     "http",
			Route:      "/myservice",
			TargetPath: "/",
		},
		Nodes: []api.Node{api.Node{api.HostPortPair{Host: "0.0", Port: 3000}}},
	}
	p.Update(services)
	_, ok := p.getServiceInfo("myservice")
	if !ok {
		t.Fatal("exptected serviceInfo to be present")
	}
}

func newProxyServer(t *testing.T, endpoint string) *tstServer {
	services := make([]api.Service, 1)

	endpoint = strings.TrimPrefix(endpoint, "http://")
	host, portStr, err := net.SplitHostPort(endpoint)
	if err != nil {
		t.Fatal(err)
	}
	port, _ := strconv.Atoi(portStr)

	services[0] = api.Service{
		Name:     "myservice",
		Protocol: "tcp",
		Frontend: api.FrontendMeta{
			Scheme:     "http",
			Route:      "/myservice",
			TargetPath: "/",
		},
		Nodes: []api.Node{api.Node{api.HostPortPair{Host: host, Port: port}}},
	}
	proxy := New()
	proxy.Update(services)
	return &tstServer{httptest.NewServer(proxy)}
}

func TestProxy(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	proxyTS := newProxyServer(t, ts.URL)
	defer proxyTS.Close()
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL+"/myservice", nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("server responded != 200")
	}
}

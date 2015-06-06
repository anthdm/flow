package proxy

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/twanies/flow/api"
)

var tcpServerPort int

func TestUpdateProxier(t *testing.T) {
	lb := NewServiceBalancer()
	service := ServicePortName{"foo", ""}
	lb.Update([]api.EndpointSet{
		{
			Name:      "foo",
			Endpoints: []api.Endpoint{{Host: "127.0.0.1", Port: tcpServerPort}},
		},
	})
	proxier := NewProxier(lb)
	waitNumLoops(t, proxier, 0)
	_, err := proxier.addServiceToPort(service, "tcp", 3000)
	if err != nil {
		t.Fatal(err)
	}
	testReadWriteTCP(t, "127.0.0.1", 3000)
	waitNumLoops(t, proxier, 1)
}

func TestUpdateDelete(t *testing.T) {
	lb := NewServiceBalancer()
	service := ServicePortName{"chat", ""}
	lb.Update([]api.EndpointSet{
		{
			Name:      "chat",
			Endpoints: []api.Endpoint{{Host: "127.0.0.1", Port: tcpServerPort}},
		},
	})
	proxier := NewProxier(lb)
	info, err := proxier.addServiceToPort(service, "tcp", 3001)
	if err != nil {
		t.Fatal(err)
	}
	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", info.proxyPort))
	if err != nil {
		t.Fatal("failed to connect to the proxy")
	}
	conn.Close()
	proxier.Update([]api.Service{})
	if _, ok := proxier.getServiceInfo(service); ok {
		t.Fatal("expected service not to be present in the serviceMap")
	}
}

func TestTcpUpdateDeleteUpdate(t *testing.T) {
	lb := NewServiceBalancer()
	service := ServicePortName{"chat", ""}
	lb.Update([]api.EndpointSet{
		{
			Name:      "chat",
			Endpoints: []api.Endpoint{{Host: "127.0.0.1", Port: tcpServerPort}},
		},
	})
	proxier := NewProxier(lb)
	info, err := proxier.addServiceToPort(service, "tcp", 3002)
	if err != nil {
		t.Fatal(err)
	}
	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", info.proxyPort))
	if err != nil {
		t.Fatalf("failed to connect to the proxy: %v", err)
	}
	conn.Close()
	proxier.Update([]api.Service{})
	if _, ok := proxier.getServiceInfo(service); ok {
		t.Fatal("expected service not to be present in the serviceMap")
	}
	proxier.Update([]api.Service{api.Service{Name: "chat", Protocol: "tcp"}})
	info, ok := proxier.getServiceInfo(service)
	if !ok {
		t.Fatal("exptected service to be present in the serviceMap")
	}
	testReadWriteTCP(t, "127.0.0.1", info.proxyPort)
}

func TestCloseProxy(t *testing.T) {
	lb := NewServiceBalancer()
	service := ServicePortName{"foo", ""}
	lb.Update([]api.EndpointSet{
		{
			Name:      "chat",
			Endpoints: []api.Endpoint{{Host: "127.0.0.1", Port: tcpServerPort}},
		},
	})
	proxier := NewProxier(lb)
	info, err := proxier.addServiceToPort(service, "tcp", 3001)
	if err != nil {
		t.Fatal(err)
	}
	proxier.Update([]api.Service{})
	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", info.proxyPort))
	if err == nil && conn != nil {
		t.Fatal("service is not stopped: proxy stil running")
	}
}

func testReadWriteTCP(t *testing.T, url string, port int) {
	resp, err := http.Get(fmt.Sprintf("http://%s:%d/%s", url, port, "foobar"))
	if err != nil {
		t.Fatalf("failed to connect to server: %v", err)
	}
	defer resp.Body.Close()
	p, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read the response body %v", err)
	}
	if !strings.Contains(string(p), "foobar") {
		t.Fatalf("expected the response body to be foobar, got %s", string(p))
	}
}

func waitNumLoops(t *testing.T, p *Proxier, want int32) {
	var got int32
	for i := 0; i < 4; i++ {
		got = atomic.LoadInt32(&p.numLoops)
		if got == want {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Errorf("expected %d ProxyLoops running, got %d", want, got)
}

func init() {
	readWriteHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(r.URL.Path))
	}
	tcpServer := httptest.NewServer(http.HandlerFunc(readWriteHandler))
	url, err := url.Parse(tcpServer.URL)
	if err != nil {
		panic(err)
	}
	_, port, _ := net.SplitHostPort(url.Host)
	tcpServerPort, err = strconv.Atoi(port)
	if err != nil {
		panic(err)
	}
}

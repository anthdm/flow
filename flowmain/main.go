package flowmain

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/twanies/flow/api/apiserver"
	"github.com/twanies/flow/pkg/proxy"
)

func Main() {
	var listen = flag.String("listen", ":9999", "")

	apiServer := apiserver.NewServer(":5001")
	apiServer.ServeAPI()

	loadBalancer := proxy.NewServiceBalancer()
	proxier := proxy.NewProxier(loadBalancer)
	go proxier.Discover()

	srv := NewServer(*listen, &reverseProxyHandler{})
	srv.CloseTimeout = 2 * time.Second
	srv.ListenAndServe()
	log.Printf("accepting work on http://localhost%s", *listen)

	srv.WaitForInterupt()
	os.Exit(0)
}

type reverseProxyHandler struct {
}

func (h *reverseProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, world"))
}

// func makeProxy() (*proxy.Proxy, error) {
// 	registry := registry.NewRegistry()
// 	services, err := registry.ListServices()
// 	if err != nil {
// 		return nil, err
// 	}
// 	p := proxy.New()
// 	p.Discover(registry)
// 	p.Update(services)
// 	return p, nil
// }

func init() {
	log.SetPrefix("flow: ")
	log.SetFlags(0)
}

package flowmain

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/twanies/flow/api"
	"github.com/twanies/flow/pkg/proxy"
)

func Main() {
	var listen = flag.String("listen", ":9999", "")

	// testing purposes
	services := make([]api.Service, 1)
	services[0] = api.Service{
		Name:     "myservice",
		Protocol: "tcp",
		Frontend: api.FrontendMeta{
			Scheme:     "http",
			Route:      "/test",
			TargetPath: "/",
		},
		Nodes: []api.Node{api.Node{api.HostPortPair{Host: "google.com", Port: 80}}},
	}

	p := proxy.New()
	p.Update(services)
	srv := NewServer(*listen, p)
	srv.CloseTimeout = 2 * time.Second
	srv.ListenAndServe()
	log.Printf("accepting work on http://localhost%s", *listen)

	srv.WaitForInterupt()
	os.Exit(0)
}

func init() {
	log.SetPrefix("flow: ")
	log.SetFlags(0)
}

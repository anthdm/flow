package main

import (
	"flag"
	"log"
	"os"
	"runtime"

	"github.com/twanies/flow/api/apiserver"
	"github.com/twanies/flow/pkg/proxy"
	"github.com/twanies/flow/pkg/watch"
)

var (
	listen       = flag.String("listen", ":9999", "")
	listenAPI    = flag.String("listenapi", ":5001", "")
	etcdMachines = flag.String("machines", "h", "")
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	os.Setenv("FLOW_MACHINES", *etcdMachines)

	apiServer := apiserver.NewServer(*listenAPI)
	apiServer.ServeAPI()

	serviceWatcher := watch.NewServiceWatcher()
	endpointWatcher := watch.NewEndpointWatcher()
	loadBalancer := proxy.NewServiceBalancer()
	proxier := proxy.NewProxier(loadBalancer)

	// register proxier and loadbalancer to the watchers so they can start
	// watching for changes and update them.
	serviceWatcher.RegisterHandler(proxier)
	endpointWatcher.RegisterHandler(loadBalancer)

	// block forever.. for now.
	select {}
}

func init() {
	log.SetPrefix("flow: ")
	log.SetFlags(0)
}

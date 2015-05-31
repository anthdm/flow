package flowmain

import (
	"flag"
	"log"
	"os"
	"time"
)

func Main() {
	var listen = flag.String("listen", ":9999", "")

	srv := NewServer(*listen, nil)
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

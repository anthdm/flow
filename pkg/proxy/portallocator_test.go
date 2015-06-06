package proxy

import (
	"log"
	"testing"
)

func TestNextPort(t *testing.T) {
	p := NewPortAllocator(2000, 3000)

	port := make(chan int, 1000)
	for x := 0; x < cap(port); x++ {
		go func() {
			port <- p.ClaimPort()
		}()
	}

	for x := 0; x < cap(port); x++ {
		p := <-port
		if p == -1 {
			t.Fatal("port allready claimed")
		}
		log.Println(p)
	}

}

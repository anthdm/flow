package proxy

import "testing"

func TestAssignFullRange(t *testing.T) {
	p := NewPortAllocator(2000, 3000)

	port := make(chan int, 1000)
	for x := 0; x < cap(port); x++ {
		go func() {
			nextPort, err := p.AssignNext()
			if err != nil {
				t.Fatal(err)
			}
			port <- nextPort
		}()
	}
	for x := 0; x < cap(port); x++ {
		p := <-port
		if p == -1 {
			t.Fatal("port allready claimed")
		}
	}
}

func TestAssignAllReleaseAll(t *testing.T) {
	portAlloc := NewPortAllocator(2000, 2500)
	port := make(chan int, 500)
	for x := 0; x < cap(port); x++ {
		go func() {
			nextPort, err := portAlloc.AssignNext()
			if err != nil {
				t.Fatal(err)
			}
			port <- nextPort
		}()
	}
	for x := 0; x < cap(port); x++ {
		p := <-port
		if p == -1 {
			t.Fatal("port allready claimed")
		}
		portAlloc.Release(p)
	}
	_, err := portAlloc.AssignNext()
	if err != nil {
		t.Fatal(err)
	}
}

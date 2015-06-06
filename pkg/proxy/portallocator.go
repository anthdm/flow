package proxy

import (
	"errors"
	"math/big"
	"math/rand"
	"sync"
)

var ErrPortClaimed = errors.New("port allready claimed")

type PortAllocator struct {
	port chan int
	min  int
	max  int

	mu      sync.RWMutex
	claimed big.Int
	last    int
}

func NewPortAllocator(min, max int) *PortAllocator {
	port := make(chan int)
	p := &PortAllocator{
		port: port,
		min:  min,
		max:  max,
	}
	go p.run(port)
	return p
}

func (p *PortAllocator) portRange() int {
	return p.max - p.min
}

func (p *PortAllocator) AssignNext() (int, error) {
	port := <-p.port
	if port == -1 {
		return -1, ErrPortClaimed
	}
	return port, nil
}

func (p *PortAllocator) run(port chan int) {
	for {
		port <- p.nextPort()
	}
}

func (p *PortAllocator) nextPort() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	portRange := p.portRange()
	i := rand.Intn(portRange)
	if b := p.claimed.Bit(i); b == 0 {
		p.claimed.SetBit(&p.claimed, i, 1)
		port := p.min + i
		p.last = port
		return port
	}
	for x := i + 1; x < portRange; x++ {
		if b := p.claimed.Bit(x); b == 0 {
			p.claimed.SetBit(&p.claimed, x, 1)
			port := p.min + x
			p.last = port
			return port
		}
	}
	for x := 0; x < i; x++ {
		if b := p.claimed.Bit(x); b == 0 {
			p.claimed.SetBit(&p.claimed, x, 1)
			port := p.min + x
			p.last = port
			return port
		}
	}
	return -1
}

func (p *PortAllocator) Release(port int) {
	port -= p.min
	if port < 0 || port > p.portRange() {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.claimed.SetBit(&p.claimed, port, 0)
}

package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// server wraps a default http.Server and hooks on the ConnState func. It allows
// to add a waitgroup on new connection and remove on hijacked or closed.
type server struct {
	*http.Server
	quit         chan bool
	fquit        chan bool
	CloseTimeout time.Duration
	wg           sync.WaitGroup
}

func NewServer(addr string, handler http.Handler) *server {
	return &server{
		Server: &http.Server{Addr: addr, Handler: handler},
		quit:   make(chan bool),
		fquit:  make(chan bool),
	}
}

func (s *server) ListenAndServe() error {
	l, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	go s.listenSignals(l)
	go s.Serve(l)
	return nil
}

func (s *server) Serve(l net.Listener) error {
	s.Server.ConnState = func(conn net.Conn, state http.ConnState) {
		switch state {
		case http.StateNew:
			s.wg.Add(1)
		case http.StateClosed, http.StateHijacked:
			s.wg.Done()
		}
	}
	return s.Server.Serve(l)
}

func (s *server) WaitForInterupt() {
	for {
		select {
		case <-s.quit:
			log.Println("receiving interupt closing all connections..")
			if s.CloseTimeout > 0 {
				time.Sleep(s.CloseTimeout)
				return
			}
			s.wg.Wait()
			return
		case <-s.fquit:
			log.Println("stopped")
			return
		}
	}
}

// listen for signals given to the server and determine its a gracefull quit or
// a force quit
func (s *server) listenSignals(l net.Listener) {
	sig := make(chan os.Signal, 1)

	signal.Notify(
		sig,
		syscall.SIGTERM,
		syscall.SIGKILL,
		syscall.SIGQUIT,
		syscall.SIGUSR2,
		syscall.SIGINT,
	)
	sign := <-sig
	switch sign {
	case syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT:
		l.Close()
		s.quit <- true
	case syscall.SIGKILL:
		l.Close()
		s.fquit <- true
	case syscall.SIGUSR2:
		log.Println("no implementation of USR2 signals")
	}
}

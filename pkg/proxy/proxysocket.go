package proxy

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"
)

const (
	RWBufferSize = 32 << 10

	// std error when tryin to use a closed network connection
	useCloseConn = "use of closed network connection"
)

type ProxySocket interface {
	ProxyLoop(service ServicePortName, info *serviceInfo, proxy *Proxier)
	Close() error
}

func newProxySocket(protocol string, port int) (ProxySocket, error) {
	switch strings.ToUpper(protocol) {
	case "TCP":
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			return nil, err
		}
		return &tcpSocket{listener}, nil
	default:
		return nil, fmt.Errorf("no implementation for %s", protocol)
	}
}

type tcpSocket struct {
	net.Listener
}

// connect attemps to connect to the destination service port
// TODO: implement couple retries with incrementing timeout duration
func (tcp *tcpSocket) connect(service ServicePortName, protocol string, proxy *Proxier) (net.Conn, error) {
	endpoint, err := proxy.loadBalancer.NextEndpoint(service)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTimeout(protocol, endpoint, 2*time.Second)
	if err != nil {
		return nil, fmt.Errorf("dial failed: %v", err)
	}
	return conn, nil
}

func (tcp *tcpSocket) Close() error {
	return tcp.Listener.Close()
}

func (tcp *tcpSocket) ProxyLoop(service ServicePortName, newInfo *serviceInfo, proxy *Proxier) {
	for {
		if info, exists := proxy.getServiceInfo(service); !exists || newInfo != info {
			return // this means the old port is replaced or closed
		}
		rwc, err := tcp.Accept()
		if err != nil {
			if strings.Contains(err.Error(), useCloseConn) {
				return
			}
			log.Printf("failed to accept: %v", err)
			continue
		}
		rwr, err := tcp.connect(service, newInfo.protocol, proxy)
		if err != nil {
			log.Printf("failed to connect to service endpoint: %v", err)
			rwr.Close()
			continue
		}
		go func() {
			done := make(chan bool, 1)
			go copyContent(rwc, rwr, done)
			go copyContent(rwr, rwc, done)
			<-done
			rwc.Close()
			rwr.Close()
		}()
	}
}

func copyContent(src io.Reader, dst io.Writer, done chan bool) {
	buf := make([]byte, RWBufferSize)
	for {
		n, err := src.Read(buf)
		if err != nil {
			done <- true
			return
		}
		_, err = dst.Write(buf[:n])
		if err != nil {
			done <- true
			return
		}
	}
}

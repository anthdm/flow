package proxy

import (
	"io"
	"log"
	"net"
	"strings"
)

const (
	RWBufferSize = 32 << 10

	// std error when tryin to use a closed network connection
	useCloseConn = "use of closed network connection"
)

type ProxySocket interface {
}

type tcpSocket struct {
	net.Listener
}

// connect attemps to connect to the destination service port
func (tcp *tcpSocket) connect(service ServicePortName, proxy *Proxier) (net.Conn, error) {
	endpoint, err := proxy.loadBalancer.NextEndpoint(service)
	if err != nil {
		return nil, err
	}
	// TODO: serviceInfo protocol
	conn, err := net.Dial("TCP", endpoint)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (tcp *tcpSocket) proxyLoop(service ServicePortName, oldInfo *serviceInfo, proxy *Proxier) {
	for {
		if info, exists := proxy.getServiceInfo(service); !exists || oldInfo != info {
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
		rwr, err := tcp.connect(service, proxy)
		if err != nil {
			log.Printf("failed to connect to service endpoint: %v", err)
			rwr.Close()
			continue
		}

		done := make(chan bool, 1)
		go copyContent(rwc, rwr, done)
		go copyContent(rwr, rwc, done)
		<-done

		rwc.Close()
		rwr.Close()
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

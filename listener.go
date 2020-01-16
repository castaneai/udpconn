package udpconn

import (
	"net"
)

type Listener struct {
	lconn *net.UDPConn
}

func Listen(addr string) (*Listener, error) {
	laddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	lconn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		return nil, err
	}
	return &Listener{lconn:lconn}, nil
}

func (l *Listener) Accept() (net.Conn, error) {
	panic("implement me")
}

func (l *Listener) Close() error {
	panic("implement me")
}

func (l *Listener) Addr() net.Addr {
	panic("implement me")
}


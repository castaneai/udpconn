package udpconn

import (
	"context"
	"log"
	"net"
	"sync"
)

type Listener struct {
	ctx    context.Context
	cancel context.CancelFunc

	lconn    *net.UDPConn
	conns    map[string]*Conn
	mu       sync.RWMutex
	acceptch chan *Conn
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
	ctx, cancel := context.WithCancel(context.Background())
	l := &Listener{
		ctx:      ctx,
		cancel:   cancel,
		lconn:    lconn,
		conns:    make(map[string]*Conn),
		mu:       sync.RWMutex{},
		acceptch: make(chan *Conn),
	}
	go l.serveUDP()
	return l, nil
}

func (l *Listener) Accept() (net.Conn, error) {
	select {
	case conn := <-l.acceptch:
		return conn, nil
	case <-l.ctx.Done():
		return nil, l.ctx.Err()
	}
}

func (l *Listener) Close() error {
	l.cancel()
	return l.lconn.Close()
}

func (l *Listener) Addr() net.Addr {
	return l.lconn.LocalAddr()
}

func (l *Listener) serveUDP() {
	buf := make([]byte, 1200)
	for {
		n, raddr, err := l.lconn.ReadFrom(buf)
		if err != nil {
			select {
			case <-l.ctx.Done():
				return
			default:
			}
			// TODO: error handling on recvfrom
			log.Printf("recvfrom err: %+v", err)
			continue
		}
		conn, ok := l.getConn(raddr)
		if !ok {
			// Accept by first packet
			conn = l.addConn(raddr)
			l.acceptch <- conn
		}
		select {
		case b := <-conn.readreq:
			copy(b, buf[:n])
			conn.readres <- n
		case <-l.ctx.Done():
			return
		}
	}
}

func (l *Listener) addConn(raddr net.Addr) *Conn {
	l.mu.Lock()
	defer l.mu.Unlock()
	closecb := func(conn *Conn) {
		l.removeConn(conn)
	}
	conn := NewConn(l.ctx, closecb, raddr, l.lconn)
	l.conns[raddr.String()] = conn
	return conn
}

func (l *Listener) removeConn(conn *Conn) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.conns, conn.RemoteAddr().String())
}

func (l *Listener) getConn(raddr net.Addr) (*Conn, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	conn, ok := l.conns[raddr.String()]
	return conn, ok
}

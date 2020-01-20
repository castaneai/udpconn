package udpconn

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"
)

var ErrNetClosing = errors.New("use of closed network connection")

type Conn struct {
	ctx context.Context

	raddr   net.Addr
	pconn   net.PacketConn
	readreq chan []byte
	readres chan int
	closecb func(*Conn)
	closech chan struct{}
	closed  bool
	mu      sync.RWMutex
}

// ctx: (context) parent listener notifies closed to Conn
// closecb: (callback) Conn notifies closed to parent Listener
func NewConn(ctx context.Context, closecb func(*Conn), raddr net.Addr, pconn net.PacketConn) *Conn {
	return &Conn{
		ctx:     ctx,
		raddr:   raddr,
		pconn:   pconn,
		readreq: make(chan []byte),
		readres: make(chan int),
		closecb: closecb,
		closech: make(chan struct{}),
		closed:  false,
		mu:      sync.RWMutex{},
	}
}

func (c *Conn) Read(b []byte) (int, error) {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return 0, ErrNetClosing
	}
	c.mu.RUnlock()

	c.readreq <- b
	select {
	case rn := <-c.readres:
		return rn, nil
	case <-c.ctx.Done():
		return 0, c.ctx.Err()
	case <-c.closech:
		return 0, c.opError("read", ErrNetClosing)
	}
}

func (c *Conn) Write(b []byte) (int, error) {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return 0, ErrNetClosing
	}
	c.mu.RUnlock()

	return c.pconn.WriteTo(b, c.raddr)
}

func (c *Conn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return ErrNetClosing
	}

	c.closed = true
	c.closecb(c)
	close(c.closech)
	return nil
}

func (c *Conn) LocalAddr() net.Addr {
	return c.pconn.LocalAddr()
}

func (c *Conn) RemoteAddr() net.Addr {
	return c.raddr
}

func (c *Conn) SetDeadline(t time.Time) error {
	return nil
}

func (c *Conn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *Conn) SetWriteDeadline(t time.Time) error {
	return nil
}

func (c *Conn) opError(op string, err error) *net.OpError {
	return &net.OpError{Op: op, Net: "udp", Source: c.pconn.LocalAddr(), Addr: c.RemoteAddr(), Err: err}
}

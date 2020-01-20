package udpconn

import (
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

type accept struct {
	conn net.Conn
	err  error
}

func dialAndConnect(l *testListener) (net.Conn, error) {
	c, err := l.Dial()
	if err != nil {
		return nil, err
	}
	// In UDP, Establishing connection requires at least one packet
	if _, err := c.Write([]byte("dummy")); err != nil {
		return nil, err
	}
	return c, nil
}

func TestCountConn(t *testing.T) {
	l, err := listenRandom()
	if err != nil {
		t.Fatal(err)
	}

	acceptch := make(chan *accept)
	go func() {
		for {
			conn, err := l.Accept()
			acceptch <- &accept{conn: conn, err: err}
		}
	}()

	buf := make([]byte, 100)
	if _, err := dialAndConnect(l); err != nil {
		t.Fatal(err)
	}
	ac1 := <-acceptch

	// WONTFIX: readしないと<-readreqのチャネル取り出しでブロックされてしまう
	if _, err := ac1.conn.Read(buf); err != nil {
		t.Fatal(err)
	}
	assert.Nil(t, ac1.err)
	assert.Equal(t, 1, len(l.conns))

	if _, err := dialAndConnect(l); err != nil {
		t.Fatal(err)
	}
	ac2 := <-acceptch
	assert.Nil(t, ac2.err)
	assert.Equal(t, 2, len(l.conns))

	assert.Nil(t, ac1.conn.Close())
	assert.Equal(t, 1, len(l.conns))

	// double closing expects already closed error
	assert.True(t, errors.Is(ac1.conn.Close(), ErrNetClosing))

	assert.Nil(t, ac2.conn.Close())
	assert.Equal(t, 0, len(l.conns))
}

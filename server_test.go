package udpconn

import (
	"fmt"
	"math/rand"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

func randomPort() uint16 {
	return uint16(rnd.Intn(65535-10000) + 10000)
}

func TestIdentifyConn(t *testing.T) {
	laddr := fmt.Sprintf("127.0.0.1:%d", randomPort())
	l, err := Listen(laddr)
	if err != nil {
		t.Fatal(err)
	}

	errch := make(chan error)
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				errch <- err
				return
			}
			go func() {
				buf := make([]byte, 100)
				for {
					n, err := conn.Read(buf)
					if err != nil {
						errch <- err
						return
					}
					if _, err := conn.Write(buf[:n]); err != nil {
						errch <- err
						return
					}
				}
			}()
		}
	}()

	c1, err := net.Dial("udp", laddr)
	if err != nil {
		t.Fatal(err)
	}
	c2, err := net.Dial("udp", laddr)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c1.Write([]byte("hello-c1")); err != nil {
		t.Fatal(err)
	}
	if _, err := c2.Write([]byte("hello-c2")); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 100)
	{
		n, err := c1.Read(buf)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "hello-c1", string(buf[:n]))
	}
	{
		n, err := c2.Read(buf)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "hello-c2", string(buf[:n]))
	}
	assert.Nil(t, c1.Close())
	assert.Nil(t, c2.Close())
}

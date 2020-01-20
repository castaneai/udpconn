package udpconn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIdentifyConn(t *testing.T) {
	l, err := listenRandom()
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
					// echo server
					if _, err := conn.Write(buf[:n]); err != nil {
						errch <- err
						return
					}
				}
			}()
		}
	}()

	c1, err := l.Dial()
	if err != nil {
		t.Fatal(err)
	}
	c2, err := l.Dial()
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

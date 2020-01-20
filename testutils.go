package udpconn

import (
	"fmt"
	"math/rand"
	"net"
	"time"
)

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

func randomPort() uint16 {
	return uint16(rnd.Intn(65535-10000) + 10000)
}

type testListener struct {
	*Listener
}

func (l *testListener) Dial() (net.Conn, error) {
	return net.Dial("udp", l.Listener.Addr().String())
}

func listenRandom() (*testListener, error) {
	laddr := fmt.Sprintf("127.0.0.1:%d", randomPort())
	l, err := Listen(laddr)
	if err != nil {
		return nil, err
	}
	return &testListener{Listener: l}, nil
}

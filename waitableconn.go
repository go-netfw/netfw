package netfw

import "net"

type waitableConn struct {
	net.Conn
	waitCh chan bool
}

func newWaitableConn(conn net.Conn) *waitableConn {
	return &waitableConn{
		waitCh: make(chan bool),
		Conn:   conn,
	}
}

func (w *waitableConn) Wait() {
	<-w.waitCh
}

func (w *waitableConn) Close() error {
	select {
	case w.waitCh <- true:
	default:
	}
	return w.Conn.Close()
}

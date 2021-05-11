package netfw

import "net"

// Listener provides a facade net.Listener implementation that forwards prexisting net.Conn to Accept() calls.
// Useful for bridging alrealdy established network connections into code that expects a net.Listener
type Listener struct {
	addr        net.Addr
	nonblocking bool

	incoming chan net.Conn
	close    chan bool
}

type Option func(l *Listener)

func NonBlocking() Option {
	return func(l *Listener) {
		l.nonblocking = true
	}
}

func WithAddr(addr net.Addr) Option {
	return func(l *Listener) {
		l.addr = addr
	}
}

func NewListener(opts ...Option) *Listener {
	l := Listener{
		incoming: make(chan net.Conn),
		close:    make(chan bool),
	}

	for _, opt := range opts {
		opt(&l)
	}

	return &l
}

func (f *Listener) Forward(conn net.Conn) {
	if f.nonblocking {
		f.incoming <- conn
	} else {
		wc := newWaitableConn(conn)
		go func() {
			f.incoming <- wc
		}()
		wc.Wait()
	}
}

// Accept waits for and returns the next connection to the listener.
func (f *Listener) Accept() (net.Conn, error) {
	select {
	case conn := <-f.incoming:

		return conn, nil

	case <-f.close:
		return nil, net.ErrClosed
	}
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
func (f *Listener) Close() error {
	select {
	case f.close <- true:
	default:
	}
	return nil
}

// Addr returns the listener's network address.
func (f *Listener) Addr() net.Addr {
	return f.addr
}

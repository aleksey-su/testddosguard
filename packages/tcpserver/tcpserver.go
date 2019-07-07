package tcpserver

import (
	"io"
	"log"
	"net"
	"sync"
	"time"
)

// HandleFunc type for handle function
type HandleFunc func(conn *SafeConn) error

// TCPServer Simple safe TCP server
type TCPServer struct {
	Addr        string
	HandleFunc  HandleFunc
	IdleTimeout time.Duration
	inShutdown  bool
	listener    net.Listener
	conns       map[*SafeConn]struct{}
	mu          sync.Mutex
}

// NewTCPServer Create new TCP Server, default idle time out 60 second
func NewTCPServer(host, port string, handleFunc HandleFunc) *TCPServer {
	return NewTCPServerWithIdleTimeout(host, port, 60*time.Second, handleFunc)
}

// NewTCPServerWithIdleTimeout Create new TCP Server
func NewTCPServerWithIdleTimeout(host, port string, timeout time.Duration, handleFunc HandleFunc) *TCPServer {
	return &TCPServer{
		Addr:        host + ":" + port,
		IdleTimeout: timeout,
		HandleFunc:  handleFunc,
	}
}

// ListenAndServe Start listening and
func (srv *TCPServer) ListenAndServe() error {
	var err error
	log.Printf("starting server on %v\n", srv.Addr)
	srv.listener, err = net.Listen("tcp", srv.Addr)
	if err != nil {
		return err
	}
	defer srv.listener.Close()
	for {
		if srv.inShutdown {
			break
		}
		conn, err := srv.listener.Accept()
		if err != nil {
			log.Printf("error accepting connection %v", err)
			continue
		}
		log.Printf("accepted connection from %v", conn.RemoteAddr())

		sConn := NewSafeConn(conn, srv.IdleTimeout)
		srv.trackConn(sConn)

		go func() {
			srv.HandleFunc(sConn)
			log.Printf("closing connection from %v", sConn.RemoteAddr())
			sConn.Close()
			srv.deleteConn(sConn)
		}()

	}
	return nil
}

// Shutdown graceful shutdown
func (srv *TCPServer) Shutdown() {
	srv.inShutdown = true
	log.Println("shutting down...")
	// srv.listener.Close()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			log.Printf("waiting on %v connections", len(srv.conns))
		}
		if len(srv.conns) == 0 {
			return
		}
	}
}

func (srv *TCPServer) trackConn(c *SafeConn) {
	defer srv.mu.Unlock()
	srv.mu.Lock()
	if srv.conns == nil {
		srv.conns = make(map[*SafeConn]struct{})
	}
	srv.conns[c] = struct{}{}
}

func (srv *TCPServer) deleteConn(conn *SafeConn) {
	defer srv.mu.Unlock()
	srv.mu.Lock()
	delete(srv.conns, conn)
}

// SafeConn Safe connection
type SafeConn struct {
	net.Conn
	IdleTimeout   time.Duration
	MaxReadBuffer int64
}

// NewSafeConn Create new safe connection
func NewSafeConn(conn net.Conn, timeout time.Duration) *SafeConn {
	sConn := &SafeConn{
		Conn:          conn,
		IdleTimeout:   timeout,
		MaxReadBuffer: 1024, // TODO: remove hardcode
	}
	sConn.SetDeadline(time.Now().Add(sConn.IdleTimeout))
	return sConn
}

func (c *SafeConn) Write(p []byte) (int, error) {
	c.updateDeadline()
	return c.Conn.Write(p)
}

func (c *SafeConn) Read(b []byte) (int, error) {
	c.updateDeadline()
	r := io.LimitReader(c.Conn, c.MaxReadBuffer)
	return r.Read(b)
}

// Close Close connection
func (c *SafeConn) Close() (err error) {
	err = c.Conn.Close()
	return
}

func (c *SafeConn) updateDeadline() {
	idleDeadline := time.Now().Add(c.IdleTimeout)
	c.Conn.SetDeadline(idleDeadline)
}

package daakia

import (
	"encoding/binary"
	"io"
	"net"
	"sync"
	"time"
)

type TCPClient struct {
	con     net.Conn
	buf_len []byte
	mu      sync.Mutex
}

func (c *TCPClient) Write(buf []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	binary.LittleEndian.PutUint32(c.buf_len, uint32(len(buf)))
	c.con.Write(c.buf_len)
	return c.con.Write(buf)
}

func (c *TCPClient) Close() error {
	return c.con.Close()
}

type TcpListener struct {
	Addr     string
	Routable func(io.WriteCloser) Router
}

func NewTCPListener(addr string, r func(io.WriteCloser) Router) *TcpListener {
	return &TcpListener{
		Addr:     addr,
		Routable: r,
	}
}

func (s *TcpListener) Listen() error {

	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	var tempDelay time.Duration
	for {
		con, err := ln.Accept()
		// Lifted from net/http package
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				//s.Logger.Printf("tcp: Accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return err
		}
		tempDelay = 0
		go s.handler(con)
	}

}

func (t *TcpListener) handler(con net.Conn) error {
	defer con.Close()
	payload_len := make([]byte, 4, 4)
	payload := make([]byte, 4096, 4096)
	router := t.Routable(&TCPClient{
		con:     con,
		buf_len: make([]byte, 4, 4),
	})
	for {
		_, err := io.ReadAtLeast(con, payload_len, 4)
		if err != nil {
			return err
		}
		n := int(binary.LittleEndian.Uint32(payload_len))
		if n > len(payload) {
			payload = make([]byte, n, n*2)
		}
		_, err = io.ReadAtLeast(con, payload[:n], n)
		if err != nil {
			return err
		}
		// Worker threads later?
		go router.Route(payload[:n])
	}
}

package daakia

import (
	"net"
	"time"
)

func NewTCPListener(addr string, bufSize int) *TCPListener {
	return &TCPListener{
		addr:    addr,
		bufSize: bufSize,
	}
}

type TCPListener struct {
	addr    string
	bufSize int
	next    func(Conn)
}

func (t *TCPListener) Next(next func(Conn)) {
	t.next = next
}

func (t *TCPListener) Listen() error {
	ln, err := net.Listen("tcp", t.addr)
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
		go t.next(NewConnection(con, t.bufSize))
	}
}

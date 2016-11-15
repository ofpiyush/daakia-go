package daakia

import (
	"encoding/binary"
	"io"
	"net"
	"sync"
	"time"
)

func NewTCPListener(addr string) *TCPListener {
	return &TCPListener{
		addr: addr,
	}
}
func NewTCPConnection(con net.Conn) *TCPConnection {
	return &TCPConnection{
		conn:          con,
		write_len_buf: make([]byte, 4, 4),
		read_len_buf:  make([]byte, 4, 4),
	}
}

type TCPListener struct {
	addr string
}

func (t *TCPListener) Listen(next func(Conn)) error {
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
		go next(NewTCPConnection(con))
	}
}

type TCPConnection struct {
	conn          net.Conn
	write_len_buf []byte
	read_len_buf  []byte
	write_len     int
	read_len      int
	wmu           sync.Mutex
	rmu           sync.Mutex
}

func (t *TCPConnection) Close() error {
	return t.conn.Close()
}

func (t *TCPConnection) Receive(payload *[]byte) (n int, err error) {
	t.rmu.Lock()
	defer t.rmu.Unlock()
	_, err = io.ReadAtLeast(t.conn, t.read_len_buf, 4)
	if err != nil {
		return
	}
	n = int(binary.LittleEndian.Uint32(t.read_len_buf))
	if len(*payload) < n {
		*payload = make([]byte, n, n)
	}
	_, err = io.ReadAtLeast(t.conn, (*payload)[:n], n)
	if err != nil {
		n = 0
		return
	}
	return
}

func (t *TCPConnection) Send(payload ...[]byte) (err error) {
	t.wmu.Lock()
	defer t.wmu.Unlock()
	t.write_len = 0
	for _, p := range payload {
		t.write_len += len(p)
	}
	binary.LittleEndian.PutUint32(t.write_len_buf, uint32(t.write_len))
	_, err = t.conn.Write(t.write_len_buf)
	if err != nil {
		return
	}
	for _, p := range payload {
		_, err = t.conn.Write(p)
		if err != nil {
			return
		}
	}
	return
}

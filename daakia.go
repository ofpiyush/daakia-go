package daakia

import (
	"io"
	"errors"
	"encoding/binary"
	"net"
	"bufio"
	"sync"
)

const (
	MAX_PAYLOAD_SIZE = 10*1024*1024
)

var (
	ErrPayloadExceeded = errors.New("Payload Exceeded max size of 10MB.")
)

type Conn interface {
	io.Closer
	Receive(*[]byte) (int, error)
	Send([]byte,[]byte) error
}

type Listener interface {
	Listen(func(Conn)) error
}

type Middleware interface {
	Next([]byte) error
}

type HttpRoutes map[string]*MethodSignature

type MethodSignature struct {
	Identifier byte
	HasPayload bool
}


func NewConnection(con net.Conn,bufSize int) *Connection {
	return &Connection{
		conn: con,
		r: bufio.NewReaderSize(con,bufSize),
		w: bufio.NewWriterSize(con,bufSize),
		write_len_buf: make([]byte, 4, 4),
		read_len_buf:  make([]byte, 4, 4),
	}
}

type Connection struct {
	conn          net.Conn
	r             *bufio.Reader
	w             *bufio.Writer
	write_len_buf []byte
	read_len_buf  []byte
	write_len     int
	read_len      int
	wmu           sync.Mutex
	rmu           sync.Mutex
}

func (c *Connection) Close() error {
	c.wmu.Lock()
	defer c.wmu.Unlock()
	c.w.Flush()
	return c.conn.Close()
}

func (c *Connection) Receive(payload *[]byte) (n int, err error) {
	//c.rmu.Lock()
	//defer c.rmu.Unlock()
	_, err = io.ReadAtLeast(c.r, c.read_len_buf, 4)
	if err != nil {
		return
	}
	n = int(binary.LittleEndian.Uint32(c.read_len_buf))
	if n > MAX_PAYLOAD_SIZE {
		return 0, ErrPayloadExceeded
	}

	if len(*payload) < n {
		*payload = make([]byte, n, n)
	}
	_, err = io.ReadAtLeast(c.r, (*payload)[:n], n)

	if err != nil {
		n = 0
		return
	}

	return
}

func (c *Connection) Send(header, payload []byte) (err error) {
	//c.wmu.Lock()
	//defer c.wmu.Unlock()
	c.write_len = len(header) + len(payload)
	if c.write_len > MAX_PAYLOAD_SIZE {
		return ErrPayloadExceeded
	}
	binary.LittleEndian.PutUint32(c.write_len_buf, uint32(c.write_len))
	_, err = c.w.Write(c.write_len_buf)
	if err != nil {
		return
	}
	_, err = c.w.Write(header)
		if err != nil {
			return
		}
	_, err = c.w.Write(payload)
		if err != nil {
			return
		}
	return
}

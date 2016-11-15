package daakia

import "io"

type Conn interface {
	io.Closer
	Receive(*[]byte) (int, error)
	Send(...[]byte) error
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

package daakia

import (
	"encoding/binary"
	"io"
	"math"
	"sync/atomic"
)

type Payload struct {
	Buf []byte

	hasReqId  bool
	hasData   bool
	dataStart int
}

func (p *Payload) Init(buf []byte) {
	p.Buf = buf

	// optimize later
	// 4 bytes of request id and one for method
	p.dataStart = 5
}
func (p *Payload) Head() int {
	return p.dataStart
}
func (p *Payload) Method() byte {
	return p.Buf[0]
}

func (p *Payload) MutMethod(b byte) {
	p.Buf[0] = b
}

func (p *Payload) ReqId() []byte {
	return p.Buf[1:p.dataStart]
}

// If you mutate the request id on a response to something random,
// It will not reach the client
// This method exists mostly to enable client call request ids to be generated
func (p *Payload) MutReqId(reqId []byte) {
	copy(p.Buf[1:p.dataStart], reqId[:4])
}

func (p *Payload) Data() []byte {
	return p.Buf[p.dataStart:]
}

// In your buffer, keep the first Payload.Head() values empty.
// Those will be overwritten
// Modeling it this way keeps it somewhat efficient.
func (p *Payload) Mutate(buf []byte) {
	copy(buf[:p.dataStart], p.Buf[:p.dataStart])
	p.Buf = buf
}

type Listener interface {
	Listen() error
}

type Router interface {
	Route([]byte) error
}

type HttpRoutes map[string]*MethodSignature

type MethodSignature struct {
	Identifier        byte
	HasPayload        bool
	HasResponse       bool
	IsServerStreaming bool
}

type RequestId struct {
	count uint32
}

func (r *RequestId) Next() uint32 {
	// Hope that someone does not have 4 gig requests in flight
	if r.count == math.MaxUint32 {
		atomic.CompareAndSwapUint32(&r.count, r.count, 0)
	}
	atomic.AddUint32(&r.count, 1)
	return r.count
}

type Connection struct {
	ReqIdGen      RequestId
	Conn          io.WriteCloser
	reqIdChannels map[byte]map[uint32]chan *Payload
}

func (c *Connection) Attach(client byte, id uint32) chan *Payload {
	if c.reqIdChannels == nil {
		c.reqIdChannels = make(map[byte]map[uint32]chan *Payload)
	}
	if c.reqIdChannels[client] == nil {
		c.reqIdChannels[client] = make(map[uint32]chan *Payload)
	}
	c.reqIdChannels[client][id] = make(chan *Payload)
	return c.reqIdChannels[client][id]
}

func (c *Connection) ClientResponse(p *Payload) {
	id := binary.LittleEndian.Uint32(p.ReqId())
	if c.reqIdChannels[p.Method()][id] == nil {
		return
	}
	c.reqIdChannels[p.Method()][id] <- p
	close(c.reqIdChannels[p.Method()][id])
	// Is this required?
	c.reqIdChannels[p.Method()][id] = nil
}

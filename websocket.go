package daakia

import (
	"golang.org/x/net/websocket"
	"net/http"
)

func NewWebSocketListener(addr string, bufSize int, mux *http.ServeMux) *WebSocketListener {
	if mux == nil {
		mux = http.NewServeMux()
	}
	return &WebSocketListener{
		addr: addr,
		mx:   mux,
		bufSize: bufSize,
	}
}

type WebSocketListener struct {
	addr string
	mx   *http.ServeMux
	bufSize int
}

func (w *WebSocketListener) Listen(next func(Conn)) error {
	w.mx.Handle("/", websocket.Handler(func(conn *websocket.Conn) {
		conn.PayloadType = websocket.BinaryFrame
		next(NewConnection(conn,w.bufSize))
	}))
	return http.ListenAndServe(w.addr, w.mx)
}

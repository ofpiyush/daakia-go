package daakia

import (
	"net/http"

	"golang.org/x/net/websocket"
)

func NewWebSocketListener(addr string, bufSize int, mux *http.ServeMux) *WebSocketListener {
	if mux == nil {
		mux = http.NewServeMux()
	}
	return &WebSocketListener{
		addr:    addr,
		mx:      mux,
		bufSize: bufSize,
	}
}

type WebSocketListener struct {
	addr    string
	mx      *http.ServeMux
	bufSize int
	next    func(Conn)
}

func (w *WebSocketListener) Next(next func(Conn)) {
	w.next = next
}

func (w *WebSocketListener) Listen() error {
	w.mx.Handle("/", websocket.Handler(func(conn *websocket.Conn) {
		conn.PayloadType = websocket.BinaryFrame
		w.next(NewConnection(conn, w.bufSize))
	}))
	return http.ListenAndServe(w.addr, w.mx)
}

package daakia

import (
	"golang.org/x/net/websocket"
	"net/http"
	"sync"
)

func NewWebSocketListener(addr string, mux *http.ServeMux) *WebSocketListener {
	if mux == nil {
		mux = http.NewServeMux()
	}
	return &WebSocketListener{
		addr: addr,
		mx:   mux,
	}
}

func NewWebSocketConnection(conn *websocket.Conn) *WebSocketConnection {
	conn.PayloadType = websocket.BinaryFrame
	return &WebSocketConnection{
		conn: conn,
	}
}

type WebSocketListener struct {
	addr string
	mx   *http.ServeMux
}

func (w *WebSocketListener) Listen(next func(Conn)) error {
	w.mx.Handle("/", websocket.Handler(func(conn *websocket.Conn) {
		next(NewWebSocketConnection(conn))
	}))
	return http.ListenAndServe(w.addr, w.mx)
}

type WebSocketConnection struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (ws *WebSocketConnection) Close() error {
	return ws.conn.Close()
}

func (ws *WebSocketConnection) Receive(payload *[]byte) (n int, err error) {
	err = websocket.Message.Receive(ws.conn, &payload)
	if err != nil {
		return
	}
	n = len(*payload)
	return
}

func (ws *WebSocketConnection) Send(payload ...[]byte) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	w, err := ws.conn.NewFrameWriter(websocket.BinaryFrame)
	if err != nil {
		return err
	}
	for _, p := range payload {
		_, err = w.Write(p)
		if err != nil {
			return err
		}
	}
	w.Close()
	return nil
}

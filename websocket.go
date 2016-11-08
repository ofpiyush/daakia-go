package daakia

import (
	"golang.org/x/net/websocket"
	"io"
	"net/http"
)

type WebSocketListener struct {
	Addr     string
	mx       *http.ServeMux
	Routable func(io.WriteCloser) Router
}

func NewWebSocket(addr string, router func(io.WriteCloser) Router) *WebSocketListener {
	ws := &WebSocketListener{
		mx:       http.NewServeMux(),
		Addr:     addr,
		Routable: router,
	}
	ws.mx.Handle("/", websocket.Handler(func(con *websocket.Conn) {
		con.PayloadType = websocket.BinaryFrame
		defer con.Close()
		router := ws.Routable(con)
		var data []byte
		for {
			err := websocket.Message.Receive(con, &data)
			if err != nil {
				return
			}
			// Worker threads later?
			go router.Route(data)
		}
	}))
	return ws
}

func (w *WebSocketListener) Listen() error {
	return http.ListenAndServe(w.Addr, w.mx)
}

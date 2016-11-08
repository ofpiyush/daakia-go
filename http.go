package daakia

// Todo: Implement
import (
	"fmt"
	"net/http"
)

type HttpListener struct {
	Addr   string
	mx     *http.ServeMux
	Router Router
}

func NewHTTP(addr string, router Router, routes *HttpRoutes) *HttpListener {
	hp := &HttpListener{
		mx:     http.NewServeMux(),
		Addr:   addr,
		Router: router,
	}
	for url, route := range *routes {
		fmt.Println(url, route)
	}

	return hp
}

func (w *HttpListener) Listen() error {
	return http.ListenAndServe(w.Addr, w.mx)
}

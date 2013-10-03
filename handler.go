package delta

import (
	"net/http"
)

type Handler struct {
	Server *Server
}

func (handler *Handler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	backends := handler.Server.OnSelecBackendtHandler(req)

	// connect to each backends
	for i := range backends {
		writer.Write([]byte(backends[i]))
	}
}

package delta

import (
	"net"
)

type Connection struct {
	Conn net.Conn
}

func (connection *Connection) Handle(server *Server) {
	// read and parse request

	// and inflates it as HTTP request
	req := &Request{}

	// select backends
	backends := server.OnSelecBackendtHandler(req)

	// connect to each backends
	for i := range backends {
		connection.Conn.Write([]byte(backends[i]))
	}

	connection.Conn.Close()
}

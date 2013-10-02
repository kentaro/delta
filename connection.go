package delta

import (
	"net"
)

type Connection struct {
	Conn net.Conn
}

func (connection *Connection) Handle(server *Server) {
	connection.Conn.Write([]byte("fooooo"))
	connection.Conn.Close()
}

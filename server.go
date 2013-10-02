package delta

import (
	"fmt"
	"log"
	"net"
)

type Server struct {
	Host                   string
	Port                   int
	Master                 string
	Backends               []*Backend
	Listener               net.Listener
	OnSelecBackendtHandler func(req *Request) []string
}

func (server *Server) SetHost(host string) {
	server.Host = host
}

func (server *Server) SetPort(port int) {
	server.Port = port
}

func (server *Server) AddMasterBackend(name, host string, port int) {
	server.Master = name
	server.Backends = append(server.Backends, &Backend{name, host, string(port)})
}

func (server *Server) AddBackend(name, host string, port int) {
	server.Backends = append(server.Backends, &Backend{name, host, string(port)})
}

func (server *Server) OnSelectBackend(handler func(req *Request) []string) {
	server.OnSelecBackendtHandler = handler
}

func (server *Server) HandleConnection(conn net.Conn) {
	connection := &Connection{conn}
	connection.Handle(server)
}

func (server *Server) Run() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.Host, server.Port))

	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Print("failed to accept")
			continue
		}

		go server.HandleConnection(conn)
	}
}

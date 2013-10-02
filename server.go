package delta

import (
	"log"
	"net"
	"net/http"
)

type Server struct {
	Address        string
	MasterBackends []map[string]string
	Backends       []map[string]string
	Listener       net.Listener
}

func (server *Server) SetAddress(address string) {
	server.Address = address
}

func (server *Server) AddMasterBackend(name, host string, port int) {
	server.MasterBackends = append(server.MasterBackends, map[string]string{
		"Name": name,
		"Host": host,
		"Port": string(port),
	})
}

func (server *Server) AddBackend(name, host string, port int) {
	server.Backends = append(server.Backends, map[string]string{
		"Name": name,
		"Host": host,
		"Port": string(port),
	})
}

func (server *Server) Listen(address string) {
	listener, err := net.Listen("tcp", address)

	if err != nil {
		log.Fatal(err)
	}

	server.Listener = listener
}

func (server *Server) Run() {
	server.Listen(server.Address)
	http.Serve(server.Listener, nil)
}

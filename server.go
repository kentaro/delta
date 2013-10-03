package delta

import (
	"fmt"
	"log"
	"net"
	"net/http"
)

type Server struct {
	Host                   string
	Port                   int
	Master                 string
	Backends               map[string]*Backend
	Listener               net.Listener
	OnSelecBackendtHandler func(req *http.Request) []string
}

func NewServer() *Server {
    server := &Server{}
    server.Port = 8484
    server.Backends = make(map[string]*Backend)
    return server
}

func (server *Server) AddMasterBackend(name, scheme, host string, port int) {
	server.Master = name
	server.Backends[name] = &Backend{name, scheme, host, port}
}

func (server *Server) AddBackend(name, scheme, host string, port int) {
	server.Backends[name] = &Backend{name, scheme, host, port}
}

func (server *Server) OnSelectBackend(handler func(req *http.Request) []string) {
	server.OnSelecBackendtHandler = handler
}

func (server *Server) Run() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.Host, server.Port))

	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/", &Handler{server})
	log.Fatal(http.Serve(listener, nil))
}

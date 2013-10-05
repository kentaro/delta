package delta

import (
	"fmt"
	"log"
	"net"
	"net/http"
)

type Server struct {
	Host                     string
	Port                     int
	Master                   string
	Backends                 map[string]*Backend
	OnSelecBackendtHandler   func(req *http.Request) []string
	OnMungeHeaderHandler     func(backend string, header *http.Header)
	OnBackendFinishedHandler func(map[string]*Response)
}

func NewServer() *Server {
	server := new(Server)
	server.Host = "0.0.0.0"
	server.Port = 8484
	server.Backends = make(map[string]*Backend)

	// By default, all backends will be selected
	server.OnSelectBackend(func(req *http.Request) []string {
		backends := make([]string, 0)
		for key, _ := range server.Backends {
			backends = append(backends, key)
		}
		return backends
	})

	return server
}

func (server *Server) AddMasterBackend(name, host string, port int) {
	server.Master = name
	server.Backends[name] = &Backend{name, "http", host, port}
}

func (server *Server) AddBackend(name, host string, port int) {
	server.Backends[name] = &Backend{name, "http", host, port}
}

func (server *Server) OnSelectBackend(handler func(req *http.Request) []string) {
	server.OnSelecBackendtHandler = handler
}

func (server *Server) OnMungeHeader(handler func(backend string, header *http.Header)) {
	server.OnMungeHeaderHandler = handler
}

func (server *Server) OnBackendFinished(handler func(responses map[string]*Response)) {
	server.OnBackendFinishedHandler = handler
}

func (server *Server) Run() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.Host, server.Port))

	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/", &Handler{server})
	log.Fatal(http.Serve(listener, nil))
}

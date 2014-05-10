package main

import (
	".."
	"log"
	"net/http"
	"strings"
	"time"
)

func main() {
	server := delta.NewServer("0.0.0.0", 8484)

	server.AddMasterBackend("production", "127.0.0.1", 8080)
	server.AddBackend("testing", "127.0.0.1", 8081)

	server.OnSelectBackend(func(req *http.Request) []string {
		if req.Method == "GET" {
			return []string{"production", "testing"}
		} else {
			return []string{"production"}
		}
	})

	server.OnMungeHeader(func(backend string, header *http.Header, req *http.Request) {
		if backend == "testing" {
			header.Add("X-Delta-Sandbox", "1")
			header.Add("X-Forwarded-Host", strings.Split(req.Host, ":")[0])
		}
	})

	server.OnBackendFinished(func(responses map[string]*delta.Response) {
		for backend, response := range responses {
			log.Printf("%s [%d ms]: %s", backend, (response.Elapsed / time.Millisecond), response.Data)
		}
	})

	server.Run()
}

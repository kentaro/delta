package main

import (
	"github.com/kentaro/delta"
	"net/http"
)

func main() {
	server := delta.NewServer()
	server.Host = "127.0.0.1"
	server.Port = 8484
	server.AddMasterBackend("production", "http", "127.0.0.1", 8080)
	server.AddBackend("testing", "http", "127.0.0.1", 8081)
	server.OnSelectBackend(func(req *http.Request) []string {
		if req.Method == "GET" {
			return []string{"production", "testing"}
		} else {
			return []string{"production"}
		}
	})
	server.Run()
}

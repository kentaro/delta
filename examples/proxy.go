package main

import (
	"github.com/kentaro/delta"
	"net/http"
)

func main() {
	server := new(delta.Server)
	server.SetHost("127.0.0.1")
	server.SetPort(8484)
	server.AddMasterBackend("production", "production.example.com", 8080)
	server.AddBackend("testing", "testing.example.com", 8080)
	server.OnSelectBackend(func(req *http.Request) []string {
		if req.Method == "GET" {
			return []string{"production", "testing"}
		} else {
			return []string{"production"}
		}
	})
	server.Run()
}

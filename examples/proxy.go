package main

import (
	"github.com/kentaro/delta"
)

func main() {
	server := new(delta.Server)
	server.SetAddress(":8484")
    server.AddMasterBackend("production", "production.example.com", 8080)
    server.AddBackend("testing", "testing.example.com", 8080)
	server.Run()
}


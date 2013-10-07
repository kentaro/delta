# Delta

Delta is an HTTP shadow proxy server that sits between clients and your server(s) to enable "shadow requests".

It's actually just a Go port of [Kage](https://github.com/cookpad/kage). You can consult the documentation of Kage for reasons why this software matters ;)

## Usage

```go
package main

import (
	"github.com/kentaro/delta"
	"log"
	"net/http"
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

	server.OnMungeHeader(func(backend string, header *http.Header) {
		if backend == "testing" {
			header.Add("X-Delta-Sandbox", "1")
		}
	})

	server.OnBackendFinished(func(responses map[string]*delta.Response) {
		for backend, response := range responses {
			log.Printf("%s [%d ms]: %s", backend, (response.Elapsed / time.Millisecond), response.Data)
		}
	})

	server.Run()
}
```

## See Also

  * [Kage](https://github.com/cookpad/kage)
  * [Geest](https://github.com/lestrrat/p5-Geest)
    * Perl port of Kage
  * [Gor](https://github.com/buger/gor)
    * It's a tool written in Go to deal with a similar problem

## Author

  * [Kentaro Kuribayashi](http://kentarok.org/)

## License

  * MIT http://kentaro.mit-license.org/


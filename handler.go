package delta

import (
	"io/ioutil"
	"log"
	"net/http"
)

type Handler struct {
	server *Server
}

func (handler *Handler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	backendNames := handler.server.OnSelecBackendtHandler(req)
	backendCount := len(backendNames)
	ch := make(chan *Response, backendCount)

	for i := range backendNames {
		go dispatchProxyRequest(handler.server.Backends[backendNames[i]], req, ch)
	}

	requestCount := 0

	for {
		response := <-ch
		requestCount = requestCount + 1

		if response.backend.name == handler.server.Master {
			content, err := ioutil.ReadAll(response.res.Body)
			defer response.res.Body.Close()

			if err != nil {
				log.Printf("%s", err)
			}

			writer.Write(content)
		}

		if requestCount == len(backendNames) {
			break
		}
	}
}

func dispatchProxyRequest(backend *Backend, req *http.Request, ch chan *Response) {
	proxyRequest := copyRequest(backend, req)
	client := new(http.Client)
	res, err := client.Do(proxyRequest)

	if err != nil {
		log.Println(err)
	}

	ch <- &Response{backend, res}
}

func copyRequest(backend *Backend, req *http.Request) *http.Request {
	proxyRequest, err := http.NewRequest(req.Method, backend.URL(req.URL.String()), nil)

	if err != nil {
		log.Fatal(err)
	}

	proxyRequest.Proto = req.Proto
	proxyRequest.Host = backend.HostPort()

	// Copy deeply because we may modify header later
	for key, values := range req.Header {
		for i := range values {
			proxyRequest.Header.Add(key, values[i])
		}
	}

	proxyRequest.Body = req.Body

	return proxyRequest
}

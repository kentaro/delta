package delta

import (
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type Handler struct {
	server *Server
}

func (handler *Handler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	backendNames := handler.server.OnSelecBackendtHandler(req)
	backendCount := len(backendNames)
	ch := make(chan *Response, backendCount)

	for i := range backendNames {
		backend := handler.server.Backends[backendNames[i]]
		go handler.dispatchProxyRequest(backend, req, ch)
	}

	responses := make(map[string]*Response)
	requestCount := 0

	for {
		response := <-ch
		defer response.httpResponse.Body.Close()

		requestCount = requestCount + 1
		responses[response.backend.name] = response

		data, err := ioutil.ReadAll(response.httpResponse.Body)
		if err != nil {
			log.Printf("backend: %s, message: %s", response.backend.name, err)
			break
		}

		response.Data = data

		if response.backend.name == handler.server.Master {
			writer.WriteHeader(response.httpResponse.StatusCode)
			writer.Write(data)
		}

		if requestCount >= len(backendNames) {
			break
		}
	}

	if handler.server.OnBackendFinishedHandler != nil {
		handler.server.OnBackendFinishedHandler(responses)
	}
}

func (handler *Handler) dispatchProxyRequest(backend *Backend, req *http.Request, ch chan *Response) {
	proxyRequest := handler.copyRequest(backend, req)
	client := new(http.Client)

	now := time.Now()
	res, err := client.Do(proxyRequest)
	elapsed := time.Now().Sub(now)

	if err != nil {
		log.Println(err)
	}

	ch <- &Response{backend, res, make([]byte, 0), elapsed}
}

func (handler *Handler) copyRequest(backend *Backend, req *http.Request) *http.Request {
	proxyRequest, err := http.NewRequest(req.Method, backend.URL(req.URL.String()), nil)

	if err != nil {
		log.Fatal(err)
	}

	proxyRequest.Proto = req.Proto
	proxyRequest.Host = backend.HostPort()
	proxyRequest.Body = req.Body

	// Copy deeply because we may modify header later
	for key, values := range req.Header {
		for i := range values {
			proxyRequest.Header.Add(key, values[i])
		}
	}

	if handler.server.OnMungeHeaderHandler != nil {
		handler.server.OnMungeHeaderHandler(backend.name, &proxyRequest.Header)
	}

	return proxyRequest
}

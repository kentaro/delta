package delta

import (
	"log"
	"net/http"
	"time"
)

type Handler struct {
	server *Server
}

func NewHandler(server *Server) *Handler {
	return &Handler{
		server: server,
	}
}

func (handler *Handler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	backendNames := handler.server.onSelecBackendtHandler(req)
	backendCount := len(backendNames)

	masterResponseCh := make(chan *Response, 1)
	responseCh := make(chan *Response, backendCount)

	for i := range backendNames {
		backend := handler.server.Backends[backendNames[i]]
		go handler.dispatchProxyRequest(backend, req, masterResponseCh, responseCh)
	}

	// Wait for all responses asynchronously
	go func() {
		responses := make(map[string]*Response)
		requestCount := 0

		for {
			response := <-responseCh

			requestCount = requestCount + 1
			responses[response.Backend.Name] = response

			if requestCount >= len(backendNames) {
				if handler.server.onBackendFinishedHandler != nil {
					handler.server.onBackendFinishedHandler(responses)
				}

				break
			}
		}
	}()

	// Wait for only master response in a blocking way
	response := <-masterResponseCh
	writer.WriteHeader(response.HttpResponse.StatusCode)
	writer.Write(response.Data)
}

func (handler *Handler) dispatchProxyRequest(backend *Backend, req *http.Request, masterResponseCh chan *Response, responseCh chan *Response) {
	proxyRequest := handler.copyRequest(backend, req)
	client := new(http.Client)

	now := time.Now()
	res, err := client.Do(proxyRequest)
	elapsed := time.Now().Sub(now)

	if err != nil {
		log.Println(err)
	}

	response, err := NewResponse(backend, res, elapsed)

	if err != nil {
		log.Println(err)
	}

	responseCh <- response
	if backend.IsMaster {
		masterResponseCh <- response
	}
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

	if handler.server.onMungeHeaderHandler != nil {
		handler.server.onMungeHeaderHandler(backend.Name, &proxyRequest.Header)
	}

	return proxyRequest
}

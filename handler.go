package delta

import (
	"bytes"
	"io"
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
	backendNames := handler.server.onSelectBackendHandler(req)
	backendCount := len(backendNames)

	masterResponseCh := make(chan *Response, 1)
	responseCh := make(chan *Response, backendCount)
	done := make(chan bool)

	bodies := make(map[string]io.Reader)
	if req.Body != nil {
		writers := make([]io.Writer, len(backendNames))
		for i, name := range backendNames {
			b := new(bytes.Buffer)
			writers[i] = b
			bodies[name] = b
		}
		io.Copy(io.MultiWriter(writers...), req.Body)
	}

	for _, name := range backendNames {
		backend := handler.server.Backends[name]
		go handler.dispatchProxyRequest(backend, req, bodies[name], masterResponseCh, responseCh)
	}

	// Wait for all responses asynchronously
	go func() {
		responses := make(map[string]*Response)
		requestCount := 0

		for {
			response := <-responseCh

			requestCount = requestCount + 1
			if response != nil {
				responses[response.Backend.Name] = response
			}

			if requestCount >= backendCount {
				if handler.server.onBackendFinishedHandler != nil {
					handler.server.onBackendFinishedHandler(responses)
				}

				done <- true
				break
			}
		}
	}()

	// Wait for only master response in a blocking way
	response := <-masterResponseCh
	if response == nil {
		http.Error(writer, "Internal Server Error", 500)
	} else {
		for key, values := range response.HttpResponse.Header {
			for i := range values {
				writer.Header().Add(key, values[i])
			}
		}
		writer.WriteHeader(response.HttpResponse.StatusCode)

		_, err := writer.Write(response.Data)
		if err != nil {
			log.Printf("HTTP Response Write Error: %s\n", err)
		}
	}

	<-done
}

func (handler *Handler) dispatchProxyRequest(backend *Backend, req *http.Request, body io.Reader, masterResponseCh chan *Response, responseCh chan *Response) {
	proxyRequest := handler.copyRequest(backend, req, body)
	client := new(http.Client)

	now := time.Now()
	res, err := client.Do(proxyRequest)
	elapsed := time.Now().Sub(now)

	var response *Response

	if err != nil {
		log.Printf("HTTP Request Error: %s\n", err)
		response = nil
	} else {
		response, err = NewResponse(backend, res, elapsed)
		if err != nil {
			log.Printf("HTTP Response Read Error: %s\n", err)
		}
	}

	responseCh <- response
	if backend.IsMaster {
		masterResponseCh <- response
	}
}

func (handler *Handler) copyRequest(backend *Backend, req *http.Request, body io.Reader) *http.Request {
	proxyRequest, err := http.NewRequest(req.Method, backend.URL(req.URL.String()), body)

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

	if handler.server.onMungeHeaderHandler != nil {
		handler.server.onMungeHeaderHandler(backend.Name, &proxyRequest.Header, req)
	}

	return proxyRequest
}

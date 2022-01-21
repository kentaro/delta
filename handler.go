package delta

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Handler struct {
	server                *Server
	bufferPool            *sync.Pool
	maxPooledBufferLength int
}

func NewHandler(server *Server) *Handler {
	return &Handler{
		server:                server,
		maxPooledBufferLength: 10 * 1024 * 1024,
		bufferPool: &sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}
}

func (handler *Handler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	backendNames := handler.server.onSelectBackendHandler(req)
	backendCount := len(backendNames)

	masterResponseCh := make(chan *Response, 1)
	responseCh := make(chan *Response, backendCount)
	done := make(chan bool)

	bodies := make(map[string]io.Reader)
	defer handler.dismissAll(bodies)
	if req.Body != nil {
		bodySize, hasSize := contentLength(req)
		writers := make([]io.Writer, len(backendNames))
		for i, name := range backendNames {
			b := handler.bufferPool.Get().(*bytes.Buffer)
			if hasSize {
				// Ensure buffer size to avoid unnecessary buffer growths
				b.Grow(bodySize)
			}
			writers[i] = b
			bodies[name] = b
		}
		_, _ = io.Copy(io.MultiWriter(writers...), req.Body)
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
			backendName := response.Backend.Name
			body := bodies[backendName]
			// remove reference in case we won't reuse this buffer.
			delete(bodies, backendName)
			// dismiss buffer back to the pool early to save some allocations.
			handler.dismiss(body)

			requestCount = requestCount + 1
			if response.Err != nil {
				responses[backendName] = response
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
	if response == nil || response.Err != nil {
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

func (handler *Handler) dismiss(r io.Reader) {
	if v, ok := r.(*bytes.Buffer); ok {
		// We don't want to reuse huge buffers
		if v.Len() < handler.maxPooledBufferLength {
			v.Reset()
			handler.bufferPool.Put(v)
		}
	}
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
		response = NewErrorResponse(backend, err, elapsed)
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
		handler.server.onMungeHeaderHandler(backend.Name, &proxyRequest.Header)
	}

	return proxyRequest
}

func (handler *Handler) dismissAll(bodies map[string]io.Reader) {
	for _, b := range bodies {
		handler.dismiss(b)
	}
}

func contentLength(req *http.Request) (int, bool) {
	s := req.Header.Get("Content-Length")
	if s == "" {
		return 0, false
	}

	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, false
	}

	return int(i), true
}

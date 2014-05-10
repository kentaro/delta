package delta

import (
	"fmt"
	. "github.com/r7kamura/gospel"
	"github.com/r7kamura/router"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupServer() *Server {
	server := NewServer("0.0.0.0", 8484)

	server.AddMasterBackend("production", "0.0.0.0", 18080)
	server.AddBackend("testing", "0.0.0.0", 18081)

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
		}
	})
	return server
}

func launchBackend(backend string, addr string) *httptest.ResponseRecorder {
	router := router.NewRouter()
	recorder := httptest.NewRecorder()

	router.Get("/", http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(recorder, "%s", backend)
	}))

	server := &http.Server{Addr: addr, Handler: router}
	go server.ListenAndServe()

	return recorder
}

func get(handler http.Handler, path string) *httptest.ResponseRecorder {
	return request(handler, "GET", path)
}

func request(handler http.Handler, method, path string) *httptest.ResponseRecorder {
	request, _ := http.NewRequest(method, path, nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func TestHandler(t *testing.T) {
	productionResponse := launchBackend("production", ":18080")
	testingResponse := launchBackend("testing", ":18081")
	server := setupServer()
	handler := NewHandler(server)

	Describe(t, "ServeHTTP", func() {
		Context("when request to normal path", func() {
			get(handler, "/")

			It("should dispatch a request to production", func() {
				Expect(productionResponse.Body.String()).To(Equal, "production")
			})

			It("should dispatch a request to testing", func() {
				Expect(testingResponse.Body.String()).To(Equal, "testing")
			})

		})
	})
}

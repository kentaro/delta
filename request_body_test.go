package delta

import (
	. "github.com/r7kamura/gospel"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestBody(t *testing.T) {
	launchBackend := func(addr string) {
		server := &http.Server{Addr: addr, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Delta-Backend", r.Header.Get("Delta-Backend"))
			if r.Body != nil {
				io.Copy(w, r.Body)
			}
		})}
		e := server.ListenAndServe()
		if e != nil {
			t.Error(e)
		}
	}

	go launchBackend(":18082")
	go launchBackend(":18083")

	server := NewServer("0.0.0.0", 8484)
	server.AddMasterBackend("production", "localhost", 18082)
	server.AddBackend("testing", "localhost", 18083)
	server.OnSelectBackend(func(req *http.Request) []string {
		if req.Header["Delta-Test-Enabled"] != nil {
			return []string{"production", "testing"}
		} else {
			return []string{"production"}
		}
	})
	server.OnMungeHeader(func(backend string, header *http.Header, req *http.Request) {
		header.Add("Delta-Backend", backend)
	})

	shouldquit := make(chan map[string]*Response)
	server.OnBackendFinished(func(responses map[string]*Response) {
		shouldquit <- responses
	})

	handler := NewHandler(server)

	request, _ := http.NewRequest("POST", "/", strings.NewReader("github.com/kentaro/delta"))
	request.Header.Add("Delta-Test-Enabled", "1")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	responses := <-shouldquit

	Describe(t, "ServeHTTP", func() {
		Context("when request have request body", func() {
			It("should have 2 responses", func() {
				Expect(len(responses)).To(Equal, 2)
			})
			It("should pass request body to production backend", func() {
				response, ok := responses["production"]
				Expect(ok).To(Equal, true)
				Expect(response.HttpResponse.Header.Get("Delta-Backend")).To(Equal, "production")
				Expect(string(response.Data)).To(Equal, "github.com/kentaro/delta")
			})
			It("should pass request body to testing backend", func() {
				response, ok := responses["testing"]
				Expect(ok).To(Equal, true)
				Expect(response.HttpResponse.Header.Get("Delta-Backend")).To(Equal, "testing")
				Expect(string(response.Data)).To(Equal, "github.com/kentaro/delta")
			})
			It("should return production response", func() {
				Expect(recorder.HeaderMap.Get("Delta-Backend")).To(Equal, "production")
				Expect(recorder.Body.String()).To(Equal, "github.com/kentaro/delta")
			})
		})
	})
}

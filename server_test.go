package delta

import (
	. "github.com/r7kamura/gospel"
	"net/http"
	"testing"
)

func TestServer(t *testing.T) {
	Describe(t, "NewServer", func() {
		server := NewServer("0.0.0.0", 8484)

		It("should set host into its Host field", func() {
			Expect(server.Host).To(Equal, "0.0.0.0")
		})

		It("should set port into its Port field", func() {
			Expect(server.Port).To(Equal, 8484)
		})

		It("should set a handler to select backend by default", func() {
			Expect(server.onSelectBackendHandler).To(Exist)
		})
	})

	Describe(t, "AddMasterBackend", func() {
		server := NewServer("0.0.0.0", 8484)

		It("should set a backend as a master", func() {
			count := len(server.Backends)
			server.AddMasterBackend("master", "127.0.01", 8080)

			Expect(server.Backends["master"].IsMaster).To(Equal, true)
			Expect(len(server.Backends)).To(Equal, count+1)
		})
	})

	Describe(t, "AddBackend", func() {
		server := NewServer("0.0.0.0", 8484)

		It("should set a backend as a testing server", func() {
			count := len(server.Backends)
			server.AddBackend("testing", "127.0.01", 8081)

			Expect(server.Backends["testing"].IsMaster).To(Equal, false)
			Expect(len(server.Backends)).To(Equal, count+1)
		})
	})

	Describe(t, "OnSelectBackend", func() {
		server := NewServer("0.0.0.0", 8484)

		It("should set a handler into its onSelectBackendHandler field", func() {
			server.OnSelectBackend(func(req *http.Request) []string {
				return []string{"testing"}
			})
			Expect(server.onSelectBackendHandler).To(NotEqual, nil)
		})
	})

	Describe(t, "OnMungeHeader", func() {
		server := NewServer("0.0.0.0", 8484)

		It("should set a handler into its onMungeHeaderHandler field", func() {
			server.OnMungeHeader(func(backend string, header *http.Header, req *http.Request) {})
			Expect(server.onMungeHeaderHandler).To(Exist)
		})
	})

	Describe(t, "OnBackendFinished", func() {
		server := NewServer("0.0.0.0", 8484)

		It("should set a handler into its onBackendFinishedHandler field", func() {
			server.OnBackendFinished(func(responses map[string]*Response) {})
			Expect(server.onBackendFinishedHandler).To(Exist)
		})
	})
}

package delta

import (
    . "github.com/r7kamura/gospel"
    "testing"
)

func TestServer(t *testing.T) {
    Describe(t, "NewServer", func () {
        server := NewServer("0.0.0.0", 8484)

        It("should set host into its Host field", func () {
            Expect(server.Host).To(Equal, "0.0.0.0")
        })

        It("should set port into its Port field", func () {
            Expect(server.Port).To(Equal, 8484)
        })

        It("should set a handler to select backend by default", func () {
            Expect(server.onSelecBackendtHandler).To(Exist)
        })
    })
}

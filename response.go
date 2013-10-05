package delta

import (
	"net/http"
)

type Response struct {
	backend      *Backend
	httpResponse *http.Response
	Data         []byte
}

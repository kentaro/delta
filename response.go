package delta

import (
	"net/http"
	"time"
)

type Response struct {
	backend      *Backend
	httpResponse *http.Response
	Data         []byte
	Elapsed      time.Duration
}

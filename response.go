package delta

import (
	"io/ioutil"
	"net/http"
	"time"
)

type Response struct {
	backend      *Backend
	httpResponse *http.Response
	Data         []byte
	Elapsed      time.Duration
}

func NewResponse(backend *Backend, httpResponse *http.Response, elapsed time.Duration) (*Response, error) {
	response := &Response{backend, httpResponse, make([]byte, 0), elapsed}
	data, err := ioutil.ReadAll(httpResponse.Body)

	response.httpResponse.Body.Close()
	response.Data = data

	return response, err
}

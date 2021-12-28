package delta

import (
	"io/ioutil"
	"net/http"
	"time"
)

type Response struct {
	Backend      *Backend
	HttpResponse *http.Response
	Data         []byte
	Elapsed      time.Duration
	Err          error
}

func NewResponse(backend *Backend, httpResponse *http.Response, elapsed time.Duration) (response *Response, err error) {
	response = &Response{
		Backend:      backend,
		HttpResponse: httpResponse,
		Data:         make([]byte, 0),
		Elapsed:      elapsed,
	}

	var data []byte
	data, err = ioutil.ReadAll(httpResponse.Body)
	_ = response.HttpResponse.Body.Close()
	response.Data = data

	return response, err
}

func NewErrorResponse(backend *Backend, err error, elapsed time.Duration) *Response {
	return &Response{
		Backend: backend,
		Err:     err,
		Elapsed: elapsed,
	}
}

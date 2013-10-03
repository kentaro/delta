package delta

import (
	"net/http"
)

type Response struct {
	backend *Backend
	res     *http.Response
}

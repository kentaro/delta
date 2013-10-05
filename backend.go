package delta

import (
	"fmt"
)

type Backend struct {
	name   string
	scheme string
	host   string
	port   int
}

func (backend *Backend) URL (pathQuery string) string {
    return fmt.Sprintf("%s://%s%s", backend.scheme, backend.HostPort(), pathQuery)
}

func (backend *Backend) HostPort() string {
	return fmt.Sprintf("%s:%d", backend.host, backend.port)
}


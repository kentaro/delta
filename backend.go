package delta

import (
	"fmt"
)

type Backend struct {
	IsMaster bool
	Name     string
	Host     string
	Port     int
}

func (backend *Backend) URL(pathQuery string) string {
	return fmt.Sprintf("http://%s%s", backend.HostPort(), pathQuery)
}

func (backend *Backend) HostPort() string {
	return fmt.Sprintf("%s:%d", backend.Host, backend.Port)
}

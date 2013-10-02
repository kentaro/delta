package delta

type Request struct {
}

func (req *Request) Method() string {
	return "GET"
}

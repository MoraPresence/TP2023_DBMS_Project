package resp

import "net/http"

type HttpErr struct {
	Message string `json:"message"`
}

type HttpResp struct {
	Data   interface{} `json:"data,omitempty"`
	Errors []HttpErr   `json:"errors"`
}

type Stat struct {
	http.ResponseWriter
	Status int
}

func (rec *Stat) WriteHeader(code int) {
	rec.Status = code
	rec.ResponseWriter.WriteHeader(code)
}

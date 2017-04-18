package testkit

import (
	"encoding/json"
	"net/http"
)

type ResponseWriterMock struct {
	HeaderMap http.Header
	Body      []byte
	Status    int
}

func NewResponseWriterMock() *ResponseWriterMock {
	return &ResponseWriterMock{
		HeaderMap: make(http.Header),
	}
}

func (rw *ResponseWriterMock) Header() http.Header {
	return rw.HeaderMap
}

func (rw *ResponseWriterMock) Write(b []byte) (int, error) {
	rw.Body = b
	return len(b), nil
}

func (rw *ResponseWriterMock) WriteHeader(status int) {
	rw.Status = status
}

func (rw *ResponseWriterMock) JsonDecode() map[string]interface{} {
	result := make(map[string]interface{})
	json.Unmarshal(rw.Body, &result)
	return result
}

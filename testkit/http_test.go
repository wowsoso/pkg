package testkit

import (
	"reflect"
	"testing"
)

func TestHeader(t *testing.T) {
	NewResponseWriterMock().Header()
}

func TestWrite(t *testing.T) {
	mock := NewResponseWriterMock()
	res, err := mock.Write([]byte{1, 2, 3})

	if res != 3 || err != nil {
		t.Fatalf("res should be 3 err should be nil, got: %d, %v", res, err)
	}

	if reflect.DeepEqual(mock.Body, []byte{1, 2, 3}) == false {
		t.Fatalf("Body should be equal '[1,2,3]', got: %v", mock.Body)
	}
}

func TestWriteHeader(t *testing.T) {
	mock := NewResponseWriterMock()
	mock.WriteHeader(200)

	if mock.Status != 200 {
		t.Fatalf("Status should be 200")
	}
}

func TestJsonDecode(t *testing.T) {
	mock := NewResponseWriterMock()
	mock.Body = []byte(`{"a":1, "b":2}`)
	mock.JsonDecode()
}

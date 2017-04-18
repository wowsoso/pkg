package testkit

import (
	"testing"
)

func TestGetIdleLocalPort(t *testing.T) {
	if GetIdleLocalPort("127.0.0.1", 0, 1) != 0 {
		t.Fatalf("result should be 0")
	}
	if GetIdleLocalPort("127.0.0.1", 1, 2) != 0 {
		t.Fatalf("result should be 0")
	}

	if GetIdleLocalPort("127.0.0.1", 1, 8000) == 0 {
		t.Fatalf("result should be more than 0")
	}
}

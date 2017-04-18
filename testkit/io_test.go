package testkit

import (
	"fmt"
	"testing"
)

func TestReadCloserMockMethodReadShouldBeReturnZeroAndErrorWhenReadErrorIsNotNil(t *testing.T) {
	err := fmt.Errorf("test")

	rc := NewReadCloserMock(nil, err, nil, false, false)

	res, err1 := rc.Read([]byte{})
	if res != 0 || err != err1 {
		t.Fatalf("res should be 0, err should be %v, got: res: %d, err: %s", err, res, err1)
	}
}

func TestReadCloserMockMethodReadShouldBePanicWhenReadPanicIsTrue(t *testing.T) {
	rc := NewReadCloserMock(nil, nil, nil, true, false)

	res := func() (r bool) {
		defer func() {
			err := recover()
			if err != nil {
				r = true
			}
		}()

		rc.Read([]byte{})

		return r
	}()

	if res != true {
		t.Fatalf("method Read should be panic.")
	}
}

func TestReadCloserMockMethodRead(t *testing.T) {
	rc := NewReadCloserMock(nil, nil, nil, false, false)

	rc.Read([]byte{})

	rc = NewReadCloserMock([]byte{1, 2, 3}, nil, nil, false, false)

	rc.Read([]byte{})

	rc = NewReadCloserMock([]byte{1, 2, 3}, nil, nil, false, false)

	rc.Read([]byte{4, 5, 6, 7})

}

func TestReadCloserMockMethodCloseShouldReturnErrorWhenCloseErrorIsNotNil(t *testing.T) {
	err := fmt.Errorf("test")
	rc := NewReadCloserMock(nil, nil, err, false, false)

	err1 := rc.Close()
	if err != err1 {
		t.Fatalf("err should be %v, got: err: %s", err, err1)
	}

}

func TestReadCloserMockMethodCloseShouldPanicWhenClosePanicIsTrue(t *testing.T) {
	rc := NewReadCloserMock(nil, nil, nil, false, true)

	res := func() (r bool) {
		defer func() {
			err := recover()
			if err != nil {
				r = true
			}
		}()

		rc.Close()

		return r
	}()

	if res != true {
		t.Fatalf("method Close should be panic.")
	}
}

func TestReadCloserMockMethodClose(t *testing.T) {
	rc := NewReadCloserMock(nil, nil, nil, false, false)

	rc.Close()
}

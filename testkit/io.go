package testkit

import (
	"io"
)

type ReadCloserMock struct {
	Cnt        []byte
	ReadError  error
	CloseError error
	ReadPanic  bool
	ClosePanic bool
}

func NewReadCloserMock(cnt []byte, readError, closeError error, readPanic, closePanic bool) io.ReadCloser {
	return &ReadCloserMock{
		Cnt:        cnt,
		ReadError:  readError,
		CloseError: closeError,
		ReadPanic:  readPanic,
		ClosePanic: closePanic,
	}
}

func (r *ReadCloserMock) Read(p []byte) (int, error) {
	if r.ReadError != nil {
		return 0, r.ReadError
	}

	if r.ReadPanic == true {
		panic("")
	}

	lenP := len(p)
	lenCnt := len(r.Cnt)

	copy(p, r.Cnt)

	if lenCnt >= lenP {
		r.Cnt = r.Cnt[lenP:]
		return lenP, nil
	} else {
		r.Cnt = nil
		return lenCnt, io.EOF
	}

}

func (r *ReadCloserMock) Close() error {
	if r.CloseError != nil {
		return r.CloseError
	}

	if r.ClosePanic == true {
		panic("")
	}

	return nil
}

package main

import (
	"context"
	"errors"
	"io"
	"sync"
)

type iReader struct {
	size           int
	pull           chan<- *pullRequest
	cancelledMutex sync.Mutex
	cancelled      bool
	ctx            context.Context
	cancelFunc     context.CancelFunc
}

type pullRequest struct {
	buf  []byte
	back chan<- readerResult
}

type readerResult struct {
	n   int
	err error
}

func NewIReader(ctx context.Context, reader io.Reader, size int) *iReader {
	ret := new(iReader)
	ret.size = size

	pullChan := make(chan *pullRequest)
	ret.pull = pullChan
	ret.cancelled = false
	ret.cancelledMutex = sync.Mutex{}
	ret.ctx, ret.cancelFunc = context.WithCancel(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				ret.cancelledMutex.Lock()
				ret.cancelled = true
				ret.cancelledMutex.Unlock()
				return
			case back := <-pullChan:
				n, err := reader.Read(back.buf)
				back.back <- readerResult{
					n:   n,
					err: err,
				}
			}
		}
	}()

	return ret
}

func (i *iReader) Read(p []byte) (n int, err error) {
	i.cancelledMutex.Lock()
	cancelled := i.cancelled
	i.cancelledMutex.Unlock()
	if cancelled {
		return 0, errors.New("reader is already closed")
	}
	back := make(chan readerResult)
	pullReq := &pullRequest{
		buf:  p,
		back: back,
	}
	i.pull <- pullReq
	readerRes := <-back
	return readerRes.n, readerRes.err
}

func (i *iReader) Close() error {
	i.cancelFunc()
	return nil
}

func (i *iReader) Cancel() {
	i.cancelFunc()
}

package wg

import (
	"context"
	"errors"
	"sync"
	"time"
)

type WaitGroupTimeout struct {
	sync.WaitGroup
}

var (
	ErrWgWaitTimeOut = errors.New("wait timeout")
)

func (wg *WaitGroupTimeout) WaitWithTimeout(ctx context.Context, timeout time.Duration) error {
	timeoutChan := time.After(timeout)
	waitChan := make(chan struct{})

	go func() {
		wg.Wait()
		close(waitChan)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timeoutChan:
		return ErrWgWaitTimeOut
	case <-waitChan:
		return nil
	}
}

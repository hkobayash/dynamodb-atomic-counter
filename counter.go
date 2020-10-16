package main

import (
	"context"
	"log"
	"sync/atomic"
	"time"
)

type counter struct {
	cnt *uint64
	errCnt *uint64
}

func NewCounter() *counter {
	var cnt uint64
	var errCnt uint64

	return &counter{
		cnt: &cnt,
		errCnt: &errCnt,
	}
}

func (h *counter) Watch(ctx context.Context) {
	t := time.NewTicker(time.Second)
	defer t.Stop()

	for {
		select {
		case now := <-t.C:
			prevCnt := atomic.SwapUint64(h.cnt, 0)
			prevErrCnt := atomic.SwapUint64(h.errCnt, 0)
			log.Printf(
				"[%s] cnt: %d qps, err: %d qps\n",
				now.Format("2006-01-02 15:04:05"),
				prevCnt,
				prevErrCnt,
			)
		case <-ctx.Done():
			return
		}
	}
}

func (h *counter) Increment() {
	atomic.AddUint64(h.cnt, 1)
}

func (h *counter) ErrIncrement() {
	atomic.AddUint64(h.errCnt, 1)
}

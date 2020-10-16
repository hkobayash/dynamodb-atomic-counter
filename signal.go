package main

import (
	"context"
	"os"
	"os/signal"
)

type sigH struct {
	ch chan os.Signal
	h  func(os.Signal)
}

func NewSigH(handler func(os.Signal), sigs ...os.Signal) *sigH {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, sigs...)

	return &sigH{
		ch: ch,
		h:  handler,
	}
}

func (h *sigH) Run(ctx context.Context, cancel func()) {
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return
		case sig := <-h.ch:
			h.h(sig)
			return
		}
	}
}

package writers

import (
	"context"
	"io"
	"os"
	"os/signal"
	"sync/atomic"
)

const SignalChannelSize = 10

var _ io.WriteCloser = &SignalReopen{}

type SignalReopen struct {
	handle atomic.Pointer[io.WriteCloser]
	signal os.Signal
	reopen func() io.WriteCloser
}

func NewSignalReopen(w io.WriteCloser, ctx context.Context, s os.Signal, reopen func() io.WriteCloser, errCh ...chan<- error) *SignalReopen {
	ch := make(chan os.Signal, SignalChannelSize)

	signal.Notify(ch, s)

	writer := &SignalReopen{
		handle: atomic.Pointer[io.WriteCloser]{},
		signal: s,
		reopen: reopen,
	}

	writer.handle.Store(&w)

	go func() {
		closeFile := func(h io.Closer) {
			if err := h.Close(); err != nil && len(errCh) > 0 {
				errCh[0] <- err
			}
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-ch:
				newHandle := reopen()
				old := writer.handle.Swap(&newHandle)
				closeFile(*old)
			}
		}
	}()

	return writer
}

func (w *SignalReopen) Write(data []byte) (int, error) {
	handle := w.handle.Load()

	return (*handle).Write(data)
}

func (w *SignalReopen) Close() error {
	handle := w.handle.Load()

	return (*handle).Close()
}

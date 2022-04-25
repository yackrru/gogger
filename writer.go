package gogger

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)

// LogWriter is the interface that wraps standard method of Write.
type LogWriter interface {
	Write(msg string)
}

// LogStreamWriter implements LogWriter.
// It is an asynchronous log writer with an internal buffer
// of log strings and a log output goroutine.
type LogStreamWriter struct {
	buf        *bytes.Buffer
	lock       sync.Locker
	opts       LogStreamWriterOption
	shutdownCh chan struct{}
	doneCh     chan struct{}
}

// LogStreamWriterOption is the options of LogStreamWriter.
// There is no way to configure options directly in the LogStreamWriter,
// and can only change settings through this option.
type LogStreamWriterOption struct {

	// Output is the log output destination.
	// Default is os.Stderr.
	Output *os.File

	// SyncIntervalMills is the log output interval.
	// Default is 100.
	SyncIntervalMills int64
}

// NewLogStreamWriter creates LogStreamWriter but does not start logging.
// Execute LogStreamWriter.Open is required to start logging.
func NewLogStreamWriter(opts LogStreamWriterOption) *LogStreamWriter {
	// Set default output.
	if opts.Output == nil {
		opts.Output = os.Stderr
	}

	// Set default sync interval mills.
	if opts.SyncIntervalMills == 0 {
		opts.SyncIntervalMills = 100
	}

	return &LogStreamWriter{
		buf:  &bytes.Buffer{},
		lock: &sync.Mutex{},
		opts: opts,
	}
}

func (w *LogStreamWriter) Write(msg string) {
	w.lock.Lock()
	defer w.lock.Unlock()

	fmt.Fprintln(w.buf, msg)
}

// Open starts synchronization of the output goroutine.
func (w *LogStreamWriter) Open() {
	w.doneCh = make(chan struct{})

	ctx, cancel := context.WithCancel(context.Background())

	w.shutdownCh = make(chan struct{})
	go func() {
		<-w.shutdownCh
		cancel()
	}()

	go func() {
		t := time.NewTicker(time.Duration(w.opts.SyncIntervalMills) * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				w.syncFlushBuf()
				close(w.doneCh)
				return
			case <-t.C:
				w.syncFlushBuf()
			}
		}
	}()
}

// Close terminates acceptance of logs and outputs logs accumulated in the buffer.
func (w *LogStreamWriter) Close() {
	close(w.shutdownCh)
	<-w.doneCh
}

func (w *LogStreamWriter) syncFlushBuf() {
	w.lock.Lock()
	defer w.lock.Unlock()

	w.opts.Output.Write(w.buf.Bytes())
	w.buf.Reset()
}

package gogger

import (
	"bufio"
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
// It is an asynchronous log writer with an internal FIFO queue
// (channel to be exact) of log strings and a log output goroutine.
type LogStreamWriter struct {
	syncWriter *syncWriter
	opts       LogStreamWriterOption
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

	// SyncQueueSize is the log string queue size.
	// Default is 10000.
	SyncQueueSize uint16
}

type syncWriter struct {
	bufWriter *bufio.Writer
	buf       []byte
	lock      sync.Locker
	ch        chan string
	cancel    context.CancelFunc
	isClosed  bool
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

	// Set default queue size
	if opts.SyncQueueSize == 0 {
		opts.SyncQueueSize = 10000
	}

	return &LogStreamWriter{
		syncWriter: newSyncWriter(opts.Output, opts.SyncQueueSize),
		opts:       opts,
	}
}

func (w *LogStreamWriter) Write(msg string) {
	w.syncWriter.Write(msg)
}

// Open starts synchronization of the output goroutine.
func (w *LogStreamWriter) Open() {
	sw := w.syncWriter

	ctx, cancel := context.WithCancel(context.Background())
	sw.cancel = cancel

	go func(ch <-chan string) {
		var wg sync.WaitGroup
		for msg := range ch {
			wg.Add(1)
			go func(msg string) {
				defer wg.Done()
				sw.lock.Lock()
				defer sw.lock.Unlock()
				sw.buf = append(sw.buf, []byte(fmt.Sprintln(msg))...)
			}(msg)
			wg.Wait()
		}
	}(sw.ch)

	go func() {
		t := time.NewTicker(time.Duration(w.opts.SyncIntervalMills) * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				for {
					if len(sw.ch) == 0 && len(sw.buf) == 0 {
						break
					}
					syncFlushBuf(sw)
				}
				sw.isClosed = true
				return
			case <-t.C:
				syncFlushBuf(sw)
			}
		}
	}()
}

// Close terminates acceptance of logs and outputs logs accumulated in the queue.
// However, since asynchronous goroutine is output while synchronizing,
// a timeout is set, and synchronization is terminated when the timeout is exceeded.
func (w *LogStreamWriter) Close(timeout time.Duration) {
	sw := w.syncWriter

	close(sw.ch)
	sw.cancel()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			sw.bufWriter.Write(sw.buf)
			sw.bufWriter.Flush()
			return
		case <-time.After(100 * time.Millisecond):
			if sw.isClosed {
				cancel()
			}
		}
	}
}

func newSyncWriter(file *os.File, queueSize uint16) *syncWriter {
	bufWriter := bufio.NewWriter(file)
	return &syncWriter{
		bufWriter: bufWriter,
		lock:      new(sync.Mutex),
		ch:        make(chan string, queueSize),
	}
}

func (w *syncWriter) Write(msg string) {
	w.ch <- msg
}

func syncFlushBuf(sw *syncWriter) {
	sw.lock.Lock()
	defer sw.lock.Unlock()
	sw.bufWriter.Write(sw.buf)
	sw.bufWriter.Flush()
	sw.buf = []byte{}
}

package gogger

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)

type LogWriter interface {
	Write(msg string)
}

type LogStreamWriter struct {
	syncWriter *syncWriter
	opts       LogStreamWriterOption
}

type LogStreamWriterOption struct {
	Output            *os.File
	SyncIntervalMills int64
	SyncQueueSize     int16
}

type syncWriter struct {
	bufWriter *bufio.Writer
	buf       []byte
	lock      sync.Locker
	ch        chan string
	cancel    context.CancelFunc
	isClosed  bool
}

func NewLogStreamWriter(opts LogStreamWriterOption) *LogStreamWriter {
	// Set default queue size
	if opts.SyncQueueSize == 0 {
		opts.SyncQueueSize = 1000
	}

	return &LogStreamWriter{
		syncWriter: newSyncWriter(opts.Output, opts.SyncQueueSize),
		opts:       opts,
	}
}

func (w *LogStreamWriter) Write(msg string) {
	w.syncWriter.Write(msg)
}

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

func (w *LogStreamWriter) Close(timeout time.Duration) {
	sw := w.syncWriter

	close(sw.ch)
	sw.cancel()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			syncFlushBuf(sw)
			return
		case <-time.After(100 * time.Millisecond):
			if sw.isClosed {
				cancel()
			}
		}
	}
}

func newSyncWriter(file *os.File, queueSize int16) *syncWriter {
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

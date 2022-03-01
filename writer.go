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
	cancel     context.CancelFunc
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
	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel

	go func(ch <-chan string) {
		sw := w.syncWriter
		var wg sync.WaitGroup
		for msg := range ch {
			wg.Add(1)
			go func() {
				defer wg.Done()
				sw.lock.Lock()
				defer sw.lock.Unlock()
				sw.buf = append(sw.buf, []byte(fmt.Sprintln(msg))...)
			}()
		}
	}(w.syncWriter.ch)

	go func() {
		t := time.NewTicker(time.Duration(w.opts.SyncIntervalMills) * time.Millisecond)
		defer t.Stop()
		for {
			sw := w.syncWriter
			select {
			case <-ctx.Done():
				sw.lock.Lock()
				defer sw.lock.Unlock()
				for {
					if len(sw.ch) == 0 && len(sw.buf) == 0 {
						break
					}
					sw.bufWriter.Write(sw.buf)
					sw.buf = []byte{}
				}
				return
			case <-t.C:
				sw.lock.Lock()
				defer sw.lock.Unlock()
				sw.bufWriter.Write(sw.buf)
				sw.buf = []byte{}
			}
		}
	}()
}

func (w *LogStreamWriter) Close() {
	close(w.syncWriter.ch)
	w.cancel()
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

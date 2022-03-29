package gogger_test

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/yackrru/gogger"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestLogStreamWriter(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	writer := gogger.NewLogStreamWriter(gogger.LogStreamWriterOption{
		Output:        w,
		SyncQueueSize: 1000,
	})
	writer.Open()

	want := ""
	for i := 0; i < 1000; i++ {
		is := strconv.Itoa(i)
		writer.Write(is)
		want += is + "\n"
	}
	time.Sleep(1 * time.Second)
	writer.Close(1 * time.Second)
	w.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, want, buf.String())
}

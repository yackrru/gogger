package gogger

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
	"time"
)

func TestIntegration(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	opts := LogStreamWriterOption{
		Output:            w,
		SyncIntervalMills: 100,
	}
	writer := NewLogStreamWriter(opts)
	writer.Open()

	conf := &LogConfig{
		Writers:     []LogWriter{writer},
		Formatter:   NewLogSimpleFormatter(DefaultLogSimpleFormatterTmpl),
		TimeFormat:  DefaultTimeFormat,
		LogMinLevel: LevelInfo,
	}
	logger := NewLog(conf)

	logger.Info("This is the test 1.")
	writer.Close(3 * time.Second)
	w.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatal(err)
	}

	resultParts := strings.Split(buf.String(), " --- ")

	assert.True(t, strings.HasSuffix(resultParts[0], "INFO"))
	timeStr := strings.TrimRight(resultParts[0], "  INFO")
	if _, err := time.Parse(DefaultTimeFormat, timeStr); err != nil {
		t.Error("Got unexpected err: ", err)
	}

	pkgAndLog := strings.Split(resultParts[1], "] ")
	pkg := strings.TrimLeft(pkgAndLog[0], "[")
	assert.True(t, strings.HasPrefix(pkg, "gogger/logger_test.go:"))
	assert.Equal(t, "This is the test 1.\n", pkgAndLog[1])
}

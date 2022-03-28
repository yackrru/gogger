package gogger_test

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/uniplaces/carbon"
	"github.com/yackrru/gogger"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestIntegration(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	opts := gogger.LogStreamWriterOption{
		Output:            w,
		SyncIntervalMills: 100,
	}
	writer := gogger.NewLogStreamWriter(opts)
	writer.Open()

	conf := &gogger.LogConfig{
		Writers:     []gogger.LogWriter{writer},
		Formatter:   gogger.NewLogSimpleFormatter(gogger.DefaultLogSimpleFormatterTmpl),
		TimeFormat:  gogger.DefaultTimeFormat,
		LogMinLevel: gogger.LevelInfo,
	}
	logger := gogger.NewLog(conf)

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
	if _, err := time.Parse(gogger.DefaultTimeFormat, timeStr); err != nil {
		t.Error("Got unexpected err: ", err)
	}

	pkgAndLog := strings.Split(resultParts[1], "] ")
	pkg := strings.TrimLeft(pkgAndLog[0], "[")
	assert.True(t, strings.HasPrefix(pkg, "gogger/logger_test.go:"))
	assert.Equal(t, "This is the test 1.\n", pkgAndLog[1])
}

const (
	debugLogStr = "debug log"
	infoLogStr  = "info log"
	warnLogStr  = "warn log"
	errorLogStr = "error log"
)

// TestLogLevelWriter is the writer that record specific log string.
type TestLogLevelWriter struct {
	hasDebug bool
	hasInfo  bool
	hasWarn  bool
	hasError bool
}

func (w *TestLogLevelWriter) Write(msg string) {
	switch msg {
	case debugLogStr:
		w.hasDebug = true
	case infoLogStr:
		w.hasInfo = true
	case warnLogStr:
		w.hasWarn = true
	case errorLogStr:
		w.hasError = true
	}
}

// TestNonFormatter is the formatter that return raw args string.
type TestNonFormatter struct{}

func (t TestNonFormatter) Format(timestamp, level, pkg string, args ...interface{}) string {
	return fmt.Sprint(args...)
}

func TestLogLevel(t *testing.T) {
	t.Run("Debug", func(t *testing.T) {
		writer := new(TestLogLevelWriter)
		logger := createLogLevelLogger(gogger.LevelDebug, writer)
		outputLog(logger)

		assert.True(t, writer.hasDebug)
		assert.True(t, writer.hasInfo)
		assert.True(t, writer.hasWarn)
		assert.True(t, writer.hasError)
	})

	t.Run("Info", func(t *testing.T) {
		writer := new(TestLogLevelWriter)
		logger := createLogLevelLogger(gogger.LevelInfo, writer)
		outputLog(logger)

		assert.False(t, writer.hasDebug)
		assert.True(t, writer.hasInfo)
		assert.True(t, writer.hasWarn)
		assert.True(t, writer.hasError)
	})

	t.Run("Warn", func(t *testing.T) {
		writer := new(TestLogLevelWriter)
		logger := createLogLevelLogger(gogger.LevelWarn, writer)
		outputLog(logger)

		assert.False(t, writer.hasDebug)
		assert.False(t, writer.hasInfo)
		assert.True(t, writer.hasWarn)
		assert.True(t, writer.hasError)
	})

	t.Run("Error", func(t *testing.T) {
		writer := new(TestLogLevelWriter)
		logger := createLogLevelLogger(gogger.LevelError, writer)
		outputLog(logger)

		assert.False(t, writer.hasDebug)
		assert.False(t, writer.hasInfo)
		assert.False(t, writer.hasWarn)
		assert.True(t, writer.hasError)
	})

	t.Run("Off", func(t *testing.T) {
		writer := new(TestLogLevelWriter)
		logger := createLogLevelLogger(gogger.LevelOff, writer)
		outputLog(logger)

		assert.False(t, writer.hasDebug)
		assert.False(t, writer.hasInfo)
		assert.False(t, writer.hasWarn)
		assert.False(t, writer.hasError)
	})
}

func createLogLevelLogger(level gogger.LogLevel, writer gogger.LogWriter) *gogger.Log {
	conf := &gogger.LogConfig{
		Writers:     []gogger.LogWriter{writer},
		Formatter:   new(TestNonFormatter),
		TimeFormat:  gogger.DefaultTimeFormat,
		LogMinLevel: level,
	}
	return gogger.NewLog(conf)
}

func outputLog(logger gogger.Logger) {
	logger.Debug(debugLogStr)
	logger.Info(infoLogStr)
	logger.Warn(warnLogStr)
	logger.Error(errorLogStr)
}

func TestGetLogLevelStr(t *testing.T) {
	assert.Equal(t, "DEBUG", gogger.ExportGetLogLevelStr(gogger.LevelDebug))
	assert.Equal(t, "INFO", gogger.ExportGetLogLevelStr(gogger.LevelInfo))
	assert.Equal(t, "WARN", gogger.ExportGetLogLevelStr(gogger.LevelWarn))
	assert.Equal(t, "ERROR", gogger.ExportGetLogLevelStr(gogger.LevelError))
}

func TestFormatFileName(t *testing.T) {
	if _, file, _, ok := runtime.Caller(0); ok {
		assert.Equal(t, "gogger/logger_test.go", gogger.ExportFormatFileName(file))
	}
}

func TestClockNow(t *testing.T) {
	// 2017-03-08T01:24:34+00:00
	timestamp := int64(1488936274)
	timeToFreeze, err := carbon.CreateFromTimestampUTC(timestamp)
	if err != nil {
		t.Fatal(err)
	}
	carbon.Freeze(timeToFreeze.Time)
	defer carbon.UnFreeze()

	t.Run("UTC default format", func(t *testing.T) {
		clockUTC := gogger.ExportClock{
			Format: gogger.DefaultTimeFormat,
			Locale: time.UTC,
		}
		assert.Equal(t, "2017-03-08 01:24:34.000", clockUTC.Now())
	})

	t.Run("UTC custom format", func(t *testing.T) {
		clockUTC := gogger.ExportClock{
			Format: "2006/01/02 15:04:05",
			Locale: time.UTC,
		}
		assert.Equal(t, "2017/03/08 01:24:34", clockUTC.Now())
	})

	t.Run("JST default format", func(t *testing.T) {
		jst, err := time.LoadLocation("Asia/Tokyo")
		if err != nil {
			t.Fatal(err)
		}
		clockJST := gogger.ExportClock{
			Format: gogger.DefaultTimeFormat,
			Locale: jst,
		}
		assert.Equal(t, "2017-03-08 10:24:34.000", clockJST.Now())
	})
}

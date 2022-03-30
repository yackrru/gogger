package gogger

import (
	"fmt"
	"github.com/uniplaces/carbon"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type LogLevel uint8

const (
	// LevelDefault equals the LevelInfo.
	LevelDefault LogLevel = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelOff
)

const (
	DefaultTimeFormat = "2006-01-02 15:04:05.000"
)

var (
	_ Logger = new(Log)
)

// Logger is the interface that wraps the methods for logging.
type Logger interface {
	Info(args ...interface{})
	Debug(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})

	Infof(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// Log is the top-level logger instance.
// However, it is only an abstracted entrypoint,
// and need to set LogWriter for output and LogFormatter for formatting.
type Log struct {
	writers   []LogWriter
	formatter LogFormatter
	clock     clock
	adapter   *logFnAdapter
}

// LogConfig is the configurations of Log.
// There is no way to configure settings directly in the Log,
// and can only change settings through this config.
type LogConfig struct {

	// Multiple LogWriters can be configured to output logs to multiple destinations.
	Writers []LogWriter

	// Only one LogFormatter can be set per Log instance.
	Formatter LogFormatter

	// The default value for TimeFormat is DefaultTimeFormat.
	TimeFormat string

	// The default value for LogMinLevel is LevelInfo.
	LogMinLevel LogLevel
}

type logFnAdapter struct {
	debug logFn
	info  logFn
	warn  logFn
	error logFn
}

type logFn func(logger Log, args ...interface{})

func NewLog(conf *LogConfig) *Log {
	log := new(Log)

	log.writers = conf.Writers
	log.formatter = conf.Formatter

	log.clock = clock{
		Format: conf.TimeFormat,
		Locale: time.Local,
	}
	if log.clock.Format == "" {
		log.clock.Format = DefaultTimeFormat
	}

	if conf.LogMinLevel == LevelDefault {
		conf.LogMinLevel = LevelInfo
	}

	adapter := new(logFnAdapter)
	adapter.debug = emptyFn
	adapter.info = emptyFn
	adapter.warn = emptyFn
	adapter.error = emptyFn
	if LevelError >= conf.LogMinLevel {
		adapter.error = errorFn
	}
	if LevelWarn >= conf.LogMinLevel {
		adapter.warn = warnFn
	}
	if LevelInfo >= conf.LogMinLevel {
		adapter.info = infoFn
	}
	if LevelDebug >= conf.LogMinLevel {
		adapter.debug = debugFn
	}
	log.adapter = adapter

	return log
}

// AddWriters adds LogWriter to Log.
func (l *Log) AddWriters(ws ...LogWriter) {
	for _, w := range ws {
		l.writers = append(l.writers, w)
	}
}

func (l *Log) Info(args ...interface{}) {
	l.adapter.info(*l, args...)
}

func (l *Log) Debug(args ...interface{}) {
	l.adapter.debug(*l, args...)
}

func (l *Log) Warn(args ...interface{}) {
	l.adapter.warn(*l, args...)
}

func (l *Log) Error(args ...interface{}) {
	l.adapter.error(*l, args...)
}

func (l *Log) Infof(format string, args ...interface{}) {
	l.adapter.info(*l, fmt.Sprintf(format, args...))
}

func (l *Log) Debugf(format string, args ...interface{}) {
	l.adapter.debug(*l, fmt.Sprintf(format, args...))
}

func (l *Log) Warnf(format string, args ...interface{}) {
	l.adapter.warn(*l, fmt.Sprintf(format, args...))
}

func (l *Log) Errorf(format string, args ...interface{}) {
	l.adapter.error(*l, fmt.Sprintf(format, args...))
}

func infoFn(logger Log, args ...interface{}) {
	log(logger, LevelInfo, args...)
}

func debugFn(logger Log, args ...interface{}) {
	log(logger, LevelDebug, args...)
}

func warnFn(logger Log, args ...interface{}) {
	log(logger, LevelWarn, args...)
}

func errorFn(logger Log, args ...interface{}) {
	log(logger, LevelError, args...)
}

// do nothing.
func emptyFn(logger Log, args ...interface{}) {
}

func log(logger Log, level LogLevel, args ...interface{}) {
	pkg := ""
	if _, file, line, ok := runtime.Caller(3); ok {
		pkg = formatFileName(file) + ":" + strconv.Itoa(line)
	}
	msg := logger.formatter.Format(logger.clock.Now(),
		getLogLevelStr(level), pkg, args...)

	// Note that do not check if writer is nil or not.
	// This is to avoid unnecessarily increasing the amount of calculations.
	// If not set writer, there will be no log output.
	for _, w := range logger.writers {
		w.Write(msg)
	}
}

func getLogLevelStr(level LogLevel) (str string) {
	switch level {
	case LevelDebug:
		str = "DEBUG"
	case LevelInfo:
		str = "INFO"
	case LevelWarn:
		str = "WARN"
	case LevelError:
		str = "ERROR"
	}
	return str
}

func formatFileName(file string) string {
	components := strings.Split(file, "/")
	return strings.Join(components[len(components)-2:], "/")
}

type clock struct {
	Format string
	Locale *time.Location
}

func (c clock) Now() string {
	return carbon.Now().In(c.Locale).Format(c.Format)
}

package gogger

import (
	"github.com/uniplaces/carbon"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type LogLevel uint8

const (
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

type Logger interface {
	Info(args ...interface{})
	Debug(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
}

type Log struct {
	writers   []LogWriter
	formatter LogFormatter
	clock     clock
	adapter   *logFnAdapter
}

type LogConfig struct {
	Writers     []LogWriter
	Formatter   LogFormatter
	TimeFormat  string
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
	compnents := strings.Split(file, "/")
	return strings.Join(compnents[len(compnents)-2:], "/")
}

type clock struct {
	Format string
	Locale *time.Location
}

func (c clock) Now() string {
	return carbon.Now().In(c.Locale).Format(c.Format)
}

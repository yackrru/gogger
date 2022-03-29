package gogger

import (
	"fmt"
	"strings"
)

const (
	DefaultLogSimpleFormatterTmpl = "%timestamp%  %level% --- [%pkg%] %args%"
)

// LogFormatter is the interface that wraps the standard method of Format.
type LogFormatter interface {
	Format(timestamp, level, pkg string, args ...interface{}) string
}

// LogSimpleFormatter implements LogFormatter.
// It outputs simple flat string has params that are timestamp, level, pkg and args.
// Set each parameter in tmpl freely arranged by enclosing each parameter in % like %timestamp%.
// The default value of tmpl is DefaultLogSimpleFormatterTmpl.
// The args param is identical to the args passed in the Logger interface method.
type LogSimpleFormatter struct {
	tmpl string
}

func NewLogSimpleFormatter(tmpl string) *LogSimpleFormatter {
	f := new(LogSimpleFormatter)

	if tmpl == "" {
		f.tmpl = DefaultLogSimpleFormatterTmpl
	} else {
		f.tmpl = tmpl
	}

	return f
}

func (f *LogSimpleFormatter) Format(timestamp, level, pkg string, args ...interface{}) string {
	msg := strings.Replace(f.tmpl, "%timestamp%", timestamp, 1)
	msg = strings.Replace(msg, "%level%", level, 1)
	msg = strings.Replace(msg, "%pkg%", pkg, 1)

	argStr := fmt.Sprint(args...)
	msg = strings.Replace(msg, "%args%", argStr, 1)

	return msg
}

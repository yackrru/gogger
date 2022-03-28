package gogger

import (
	"fmt"
	"strings"
)

const (
	DefaultLogSimpleFormatterTmpl = "%timestamp%  %level% --- [%pkg%] %args%"
)

type LogFormatter interface {
	Format(timestamp, level, pkg string, args ...interface{}) string
}

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

# gogger

[![CI](https://github.com/yackrru/gogger/actions/workflows/ci.yml/badge.svg)](https://github.com/yackrru/gogger/actions/workflows/ci.yml)
[![CodeQL](https://github.com/yackrru/gogger/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/yackrru/gogger/actions/workflows/codeql-analysis.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/yackrru/gogger.svg)](https://pkg.go.dev/github.com/yackrru/gogger)

A Go library for logging.

## Overview
gogger is a logging library, providing just one `Log` struct and `Logger` interface.
This is to simplify instance creation and to ensure that the exact same approach can always be taken.  
`Log` struct is designed to be highly extensible, using the `LogWriter` interface for log output
and the `LogFormatter` interface for log formatting, which can be freely replaced.
Also, multiple `LogWriters` can be configured for `Log`,
so that the same log can be written to both stdout and a file, for example.  
Since `LogWriter` and `LogFormatter` are exported interfaces,
users are free to define and use them if they are dissatisfied with the built-in ones.
If they are useful, please give us pull request.

## Quickstart
Here is an example of using the built-in `LogStreamWriter` and `LogSimpleFormatter`.
Those details are described after this section.  
Remember that `Log` generation is done here by passing only `LogConfig` to its constructor, `NewLog`.
Then, configure `LogWriter` and `LogFormatter` in `LogConfig`.

```go
package main

import (
	"github.com/yackrru/gogger"
	"os"
	"time"
)

func main() {
	writer := gogger.NewLogStreamWriter(gogger.LogStreamWriterOption{
		Output: os.Stderr,
	})
	writer.Open()
	defer writer.Close(2 * time.Minute)
	
	logger := gogger.NewLog(&gogger.LogConfig{
		Writers: []gogger.LogWriter{writer},
		Formatter: gogger.NewLogSimpleFormatter(gogger.DefaultLogSimpleFormatterTmpl),
	})
	
	// logging
	logger.Info("log string")
}
```

## LogConfig
It is a centralized setting for `Log`.

|   option    |          default          | description                                                     |
|:-----------:|:-------------------------:|:----------------------------------------------------------------|
|   Writers   |             -             | Set a LogWriter for each output destination.                    |
|  Formatter  |             -             | Only one LogFormatter can be set.                               |
| TimeFormat  | `"2006-01-02 15:04:05.000"` | Time stamp format. Specify a time format string for the golang. |
| LogMinLevel |           INFO            | Specify the log level you want to output. Log levels are ERROR, WARN, INFO, DEBUG, and OFF. |

## Writer

### LogStreamWriter
It is an asynchronous log writer with an internal FIFO queue (channel to be exact) of log strings and a log output goroutine.
Performance is adjusted the capacity of the channel as a FIFO queue and the interval between log outputs.

```go
gogger.NewLogStreamWriter(gogger.LogStreamWriterOption{
	Output: os.Stderr,
	SyncIntervalMills: 100,
	SyncQueueSize: 50000,
})
```

| option |  default  | description                                               |
|:------:|:---------:|:----------------------------------------------------------|
| Output | os.Stderr | Specify the `*os.File` to which the logs will be output.  |
|SyncIntervalMills|    100    | The log output interval. (milliseconds)                   |
|SyncQueueSize|   10000   | The capacity of the channel as a FIFO queue. (0 - 65,535) |

## Formatter

### LogSimpleFormatter
It is a LogFormatter that formats a predefined parameter and log string in a format set as a template.
The struct itself has only one parameter, a template string called tmpl, which is passed to the constructor when the instance is created.

```go
gogger.NewLogSimpleFormatter(gogger.DefaultLogSimpleFormatterTmpl)
```

| option |                  default                   | description                                            |
|:------:|:------------------------------------------:|:-------------------------------------------------------|
|  tmpl  | `"%timestamp%  %level% --- [%pkg%] %args%"`  | The template string. Each parameter is enclosed in %. %args% is the string passed to `Log`. |

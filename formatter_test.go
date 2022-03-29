package gogger_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/yackrru/gogger"
	"testing"
)

func TestLogSimpleFormatterFormat(t *testing.T) {
	t.Run("Default format", func(t *testing.T) {
		formatter := gogger.NewLogSimpleFormatter(gogger.DefaultLogSimpleFormatterTmpl)
		result := formatter.Format(
			"2022-01-01 00:00:00.000", "INFO", "gogger/formatter_test.go",
			"log1 ", "log2 ", "log3")
		assert.Equal(t, "2022-01-01 00:00:00.000  INFO --- [gogger/formatter_test.go]"+
			" log1 log2 log3", result)
	})

	t.Run("Custom format", func(t *testing.T) {
		formatter := gogger.NewLogSimpleFormatter("%timestamp% %level% %pkg% %args%")
		result := formatter.Format(
			"2022-01-01 00:00:00.000", "INFO", "gogger/formatter_test.go",
			"log1 ", "log2 ", "log3")
		assert.Equal(t, "2022-01-01 00:00:00.000 INFO gogger/formatter_test.go"+
			" log1 log2 log3", result)
	})

	t.Run("Custom format without timestamp", func(t *testing.T) {
		formatter := gogger.NewLogSimpleFormatter("%level% %pkg% %args%")
		result := formatter.Format(
			"2022-01-01 00:00:00.000", "INFO", "gogger/formatter_test.go",
			"log1 ", "log2 ", "log3")
		assert.Equal(t, "INFO gogger/formatter_test.go log1 log2 log3", result)
	})

	t.Run("Custom format without level", func(t *testing.T) {
		formatter := gogger.NewLogSimpleFormatter("%timestamp% %pkg% %args%")
		result := formatter.Format(
			"2022-01-01 00:00:00.000", "INFO", "gogger/formatter_test.go",
			"log1 ", "log2 ", "log3")
		assert.Equal(t, "2022-01-01 00:00:00.000 gogger/formatter_test.go log1 log2 log3", result)
	})

	t.Run("Custom format without pkg", func(t *testing.T) {
		formatter := gogger.NewLogSimpleFormatter("%timestamp% %level% %args%")
		result := formatter.Format(
			"2022-01-01 00:00:00.000", "INFO", "gogger/formatter_test.go",
			"log1 ", "log2 ", "log3")
		assert.Equal(t, "2022-01-01 00:00:00.000 INFO log1 log2 log3", result)
	})
}

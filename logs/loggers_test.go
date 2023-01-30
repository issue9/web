// SPDX-License-Identifier: MIT

package logs

import (
	"bytes"
	"errors"
	"testing"

	"github.com/issue9/assert/v3"
)

func TestLogs_With(t *testing.T) {
	a := assert.New(t, false)
	buf := new(bytes.Buffer)

	l := New(&Options{
		Writer:  NewTextWriter(NanoLayout, buf),
		Caller:  true,
		Created: true,
		Levels:  AllLevels(),
	}, nil)
	a.NotNil(l)

	l.NewEntry(Error).DepthString(1, "error")
	a.Contains(buf.String(), "error").
		Contains(buf.String(), "loggers_test.go:25") // 依赖 DepthString 行号

	// Logs.With

	buf.Reset()
	ps := l.With(map[string]any{"k1": "v1"})
	a.NotNil(ps)
	ps.ERROR().String("string")
	a.Contains(buf.String(), "string").
		Contains(buf.String(), "loggers_test.go:34"). // 依赖 ERROR().String 行号
		Contains(buf.String(), "k1=v1")

	ps.NewEntry(Error).DepthError(1, errors.New("error"))
	a.Contains(buf.String(), "error").
		Contains(buf.String(), "loggers_test.go:39") // 依赖 DepthError 行号

	Destroy(ps)
}

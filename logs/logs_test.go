// SPDX-License-Identifier: MIT

package logs

import (
	"bytes"
	"io"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/localeutil"
	"github.com/issue9/term/v3/colors"
)

func TestLogs_With(t *testing.T) {
	a := assert.New(t, false)
	buf := new(bytes.Buffer)

	l, err := New(nil, &Options{
		Created:  NanoLayout,
		Handler:  NewTextHandler(buf),
		Location: true,
		Levels:   AllLevels(),
	})
	a.NotError(err).NotNil(l)

	l.NewRecord().DepthString(2, "error").
		Output(l.ERROR())
	a.Contains(buf.String(), "error").
		Contains(buf.String(), "logs_test.go:27") // 依赖 DepthString 行号

	// Logs.New

	buf.Reset()
	attrsLogs1 := l.New(map[string]any{"k1": "v1"})
	a.NotNil(attrsLogs1)
	attrsLogs1.ERROR().String("string")
	a.Contains(buf.String(), "string").
		Contains(buf.String(), "logs_test.go:37"). // 依赖 ERROR().String 行号
		Contains(buf.String(), "k1=v1")

	buf.Reset()
	attrsLogs2 := attrsLogs1.New(map[string]any{"k2": "v2"})
	attrsLogs2.DEBUG().String("DEBUG")
	a.Contains(buf.String(), "DEBUG").
		Contains(buf.String(), "logs_test.go:44"). // 依赖 DEBUG().String() 行号
		Contains(buf.String(), "k1=v1").
		Contains(buf.String(), "k2=v2")
	attrsLogs1.Free()
}

func TestNew(t *testing.T) {
	a := assert.New(t, false)

	l, err := New(nil, nil)
	a.NotError(err).NotNil(l)
	l.DEBUG().Println("test")

	textBuf := new(bytes.Buffer)
	termBuf := new(bytes.Buffer)
	infoBuf := new(bytes.Buffer)
	opt := &Options{
		Handler: NewDispatchHandler(map[Level]Handler{
			Error: NewTextHandler(textBuf),
			Warn:  NewTermHandler(termBuf, map[Level]colors.Color{Info: colors.Blue}),
			Info:  NewJSONHandler(infoBuf),
			Debug: NewJSONHandler(io.Discard),
			Trace: NewJSONHandler(io.Discard),
			Fatal: NewJSONHandler(io.Discard),
		}),
		Location: true,
		Created:  MicroLayout,
		Levels:   AllLevels(),
	}
	l, err = New(nil, opt)
	a.NotError(err).NotNil(l)

	l.ERROR().Error(localeutil.Error("scheduled job"))
	l.WARN().Printf("%s not found", localeutil.Phrase("scheduled job"))
	l.INFO().Print(localeutil.Phrase("scheduled job"))
	a.Contains(textBuf.String(), "scheduled job").
		Contains(termBuf.String(), "scheduled job not found").
		Contains(infoBuf.String(), "scheduled job")
}

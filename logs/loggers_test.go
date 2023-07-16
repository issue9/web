// SPDX-License-Identifier: MIT

package logs

import (
	"bytes"
	"errors"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/localeutil"
	"github.com/issue9/term/v3/colors"

	"github.com/issue9/web/internal/errs"
)

func TestLogs_With(t *testing.T) {
	a := assert.New(t, false)
	buf := new(bytes.Buffer)

	l, err := New(&Options{
		Handler: NewTextHandler(NanoLayout, buf),
		Caller:  true,
		Created: true,
		Levels:  AllLevels(),
	})
	a.NotError(err).NotNil(l)

	l.NewRecord(Error).DepthString(1, "error")
	a.Contains(buf.String(), "error").
		Contains(buf.String(), "loggers_test.go:29") // 依赖 DepthString 行号

	// Logs.With

	buf.Reset()
	ps := l.With(map[string]any{"k1": "v1"})
	a.NotNil(ps)
	ps.ERROR().String("string")
	a.Contains(buf.String(), "string").
		Contains(buf.String(), "loggers_test.go:38"). // 依赖 ERROR().String 行号
		Contains(buf.String(), "k1=v1")

	ps.NewRecord(Error).DepthError(1, errors.New("error"))
	a.Contains(buf.String(), "error").
		Contains(buf.String(), "loggers_test.go:43") // 依赖 DepthError 行号

	DestroyParamsLogs(ps)
}

func TestNew(t *testing.T) {
	a := assert.New(t, false)

	l, err := New(nil)
	a.NotError(err).NotNil(l)
	a.False(l.logs.IsEnable(Debug))
	l.DEBUG().Println("test")

	textBuf := new(bytes.Buffer)
	termBuf := new(bytes.Buffer)
	infoBuf := new(bytes.Buffer)
	opt := &Options{
		Handler: NewDispatchHandler(map[Level]Handler{
			Error: NewTextHandler(MicroLayout, textBuf),
			Warn:  NewTermHandler(MicroLayout, termBuf, map[Level]colors.Color{Info: colors.Blue}),
			Info:  NewJSONHandler(MicroLayout, infoBuf),
		}),
		Caller:  true,
		Created: true,
		Levels:  AllLevels(),
	}
	l, err = New(opt)
	a.NotError(err).NotNil(l)

	l.ERROR().Error(errs.NewLocaleError("scheduled job"))
	l.WARN().Printf("%s not found", localeutil.Phrase("scheduled job"))
	l.INFO().Print(localeutil.Phrase("scheduled job"))
	a.Contains(textBuf.String(), "scheduled job").
		Contains(termBuf.String(), "scheduled job not found").
		Contains(infoBuf.String(), "scheduled job")
}

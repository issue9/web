// SPDX-License-Identifier: MIT

package logs

import (
	"bytes"
	"errors"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/localeutil"
	xmsg "github.com/issue9/localeutil/message"
	"github.com/issue9/term/v3/colors"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
	"gopkg.in/yaml.v3"

	"github.com/issue9/web/internal/errs"
	"github.com/issue9/web/locales"
)

func TestLogs_With(t *testing.T) {
	a := assert.New(t, false)
	buf := new(bytes.Buffer)

	l, err := New(&Options{
		Writer:  NewTextWriter(NanoLayout, buf),
		Caller:  true,
		Created: true,
		Levels:  AllLevels(),
	}, nil)
	a.NotError(err).NotNil(l)

	l.NewEntry(Error).DepthString(1, "error")
	a.Contains(buf.String(), "error").
		Contains(buf.String(), "loggers_test.go:35") // 依赖 DepthString 行号

	// Logs.With

	buf.Reset()
	ps := l.With(map[string]any{"k1": "v1"})
	a.NotNil(ps)
	ps.ERROR().String("string")
	a.Contains(buf.String(), "string").
		Contains(buf.String(), "loggers_test.go:44"). // 依赖 ERROR().String 行号
		Contains(buf.String(), "k1=v1")

	ps.NewEntry(Error).DepthError(1, errors.New("error"))
	a.Contains(buf.String(), "error").
		Contains(buf.String(), "loggers_test.go:49") // 依赖 DepthError 行号

	Destroy(ps)
}

func TestNew(t *testing.T) {
	a := assert.New(t, false)

	l, err := New(nil, nil)
	a.NotError(err).NotNil(l)
	a.False(l.logs.IsEnable(Debug))
	l.DEBUG().Println("test")

	textBuf := new(bytes.Buffer)
	termBuf := new(bytes.Buffer)
	infoBuf := new(bytes.Buffer)
	opt := &Options{
		Writer: NewDispatchWriter(map[Level]Writer{
			Error: NewTextWriter(MicroLayout, textBuf),
			Warn:  NewTermWriter(MicroLayout, colors.Black, termBuf),
			Info:  NewJSONWriter(MicroLayout, infoBuf),
		}),
		Caller:  true,
		Created: true,
		Levels:  AllLevels(),
	}
	l, err = New(opt, nil)
	a.NotError(err).NotNil(l)

	l.ERROR().Error(errs.NewLocaleError("scheduled job"))
	l.WARN().Printf("%s not found", localeutil.Phrase("scheduled job"))
	l.INFO().Print(localeutil.Phrase("scheduled job"))
	a.Contains(textBuf.String(), "scheduled job").
		Contains(termBuf.String(), "scheduled job not found").
		Contains(infoBuf.String(), "scheduled job")

	// with Printer interface

	textBuf.Reset()
	termBuf.Reset()
	infoBuf.Reset()
	b := catalog.NewBuilder()
	msg := &xmsg.Messages{}
	a.NotError(msg.LoadFSGlob(locales.Locales, "*.yml", yaml.Unmarshal)).
		NotError(msg.Catalog(b))
	p := message.NewPrinter(language.SimplifiedChinese, message.Catalog(b))

	println(p.Sprintf("scheduled job"))

	opt = &Options{
		Writer: NewDispatchWriter(map[Level]Writer{
			Error: NewTextWriter(MicroLayout, textBuf),
			Warn:  NewTermWriter(MicroLayout, colors.Black, termBuf),
			Info:  NewTextWriter(MicroLayout, infoBuf),
		}),
		Caller:  true,
		Created: true,
		Levels:  AllLevels(),
	}
	l, err = New(opt, p)
	a.NotError(err).NotNil(l)

	l.ERROR().Error(errs.NewLocaleError("scheduled job"))
	l.WARN().Printf("%s not found", localeutil.Phrase("scheduled job"))
	l.INFO().Print(localeutil.Phrase("scheduled job"))
	a.Contains(textBuf.String(), "计划任务").
		Contains(termBuf.String(), "计划任务 不存在").
		Contains(infoBuf.String(), "计划任务")
}

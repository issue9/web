// SPDX-License-Identifier: MIT

package logs

import (
	"bytes"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/localeutil"
	"github.com/issue9/term/v3/colors"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
	"gopkg.in/yaml.v3"

	"github.com/issue9/web/internal/errs"
	"github.com/issue9/web/locales"
)

func TestNew(t *testing.T) {
	a := assert.New(t, false)

	l := New(nil, nil)
	a.NotNil(l)
	a.False(l.logs.IsEnable(Error))

	textBuf := new(bytes.Buffer)
	termBuf := new(bytes.Buffer)
	infoBuf := new(bytes.Buffer)
	opt := &Options{
		Writer: NewDispatchWriter(map[Level]Writer{
			Error: NewTextWriter(MicroLayout, textBuf),
			Warn:  NewTermWriter(MicroLayout, colors.Black, termBuf),
			Info:  NewTextWriter(MicroLayout, infoBuf),
		}),
		Caller:  true,
		Created: true,
		Levels:  AllLevels(),
	}
	l = New(opt, nil)
	a.NotNil(l)

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
	err := localeutil.LoadMessageFromFSGlob(b, locales.Locales, "*.yml", yaml.Unmarshal)
	a.NotError(err)
	p := message.NewPrinter(language.SimplifiedChinese, message.Catalog(b))

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
	l = New(opt, p)
	a.NotNil(l)

	l.ERROR().Error(errs.NewLocaleError("scheduled job"))
	l.WARN().Printf("%s not found", localeutil.Phrase("scheduled job"))
	l.INFO().Print(localeutil.Phrase("scheduled job"))
	a.Contains(textBuf.String(), "计划任务").
		Contains(termBuf.String(), "计划任务 不存在").
		Contains(infoBuf.String(), "计划任务")
}

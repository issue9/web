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

func TestPrinter(t *testing.T) {
	a := assert.New(t, false)
	textBuf := new(bytes.Buffer)
	termBuf := new(bytes.Buffer)
	l := New(NewDispatchWriter(map[Level]Writer{
		Error: NewTextWriter(MicroLayout, textBuf),
		Warn:  NewTermWriter(MicroLayout, colors.Black, termBuf),
		Info:  NewNopWriter(),
	}), true, true)
	a.NotNil(l)

	l.SetPrinter(nil)
	l.ERROR().Error(errs.NewLocaleError("scheduled job"))
	l.WARN().Printf("%s not found", "item")
	l.INFO().Print("info")
	a.Contains(textBuf.String(), "scheduled job").
		Contains(termBuf.String(), "item not found")

	// SetPrinter

	textBuf.Reset()
	termBuf.Reset()
	b := catalog.NewBuilder()
	err := localeutil.LoadMessageFromFSGlob(b, locales.Locales, "*.yml", yaml.Unmarshal)
	a.NotError(err)
	p := NewPrinter(message.NewPrinter(language.SimplifiedChinese, message.Catalog(b)))
	a.NotNil(p)

	l.SetPrinter(p)
	l.ERROR().Error(errs.NewLocaleError("scheduled job"))
	l.WARN().Printf("%s not found", "item")
	l.INFO().Print("info")
	a.Contains(textBuf.String(), "计划任务").
		Contains(termBuf.String(), "item 不存在")
}

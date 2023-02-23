// SPDX-License-Identifier: MIT

package logs

import (
	"fmt"

	"github.com/issue9/localeutil"
	"github.com/issue9/logs/v4"
	"golang.org/x/text/message"
)

// 声明带有翻译功能的日志转换接口对象
func newPrinter(p *message.Printer) logs.Printer { return &localePrinter{p: p} }

type localePrinter struct {
	p *message.Printer
}

func (p *localePrinter) Error(err error) string {
	if ls, ok := err.(localeutil.LocaleStringer); ok {
		return ls.LocaleString(p.p)
	}
	return err.Error()
}

func (p *localePrinter) String(s string) string { return p.p.Sprintf(s) }

func (p *localePrinter) Print(v ...any) string {
	return fmt.Sprint(p.localeValues(v...)...)
}

func (p *localePrinter) Println(v ...any) string {
	return fmt.Sprintln(p.localeValues(v...)...)
}

func (p *localePrinter) Printf(format string, v ...any) string {
	return p.p.Sprintf(format, p.localeValues(v...)...)
}

func (p *localePrinter) localeValues(v ...any) []any {
	if len(v) == 0 {
		return nil
	}

	vals := make([]any, 0, len(v))
	for _, val := range v {
		if ls, ok := val.(localeutil.LocaleStringer); ok {
			vals = append(vals, ls.LocaleString(p.p))
		} else {
			vals = append(vals, val)
		}
	}
	return vals
}

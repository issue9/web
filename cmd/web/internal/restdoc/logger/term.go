// SPDX-License-Identifier: MIT

package logger

import (
	"fmt"
	"io"

	"github.com/issue9/localeutil"
	"github.com/issue9/term/v3/colors"
)

var typeColors = map[Type]colors.Color{
	Unknown:   colors.Red,
	Info:      colors.Green,
	Warning:   colors.Yellow,
	Cancelled: colors.Yellow,
	ModSyntax: colors.Red,
	GoSyntax:  colors.Red,
	DocSyntax: colors.Red,
}

func BuildTermHandler(w io.Writer, p *localeutil.Printer) func(*Entry) {
	return func(e *Entry) {
		var msg string
		if l, ok := e.Msg.(localeutil.LocaleStringer); ok {
			msg = l.LocaleString(p)
		} else {
			msg = fmt.Sprint(e.Msg)
		}

		if e.Filename != "" {
			msg = localeutil.Phrase("%s at %s:%d\n", msg, e.Filename, e.Line).LocaleString(p)
		}

		colors.Fprintf(w, colors.Normal, typeColors[e.Type], colors.Default, msg)
	}
}

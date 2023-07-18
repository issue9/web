// SPDX-License-Identifier: MIT

package logger

import (
	"bytes"
	"errors"
	"go/scanner"
	"go/token"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/web/logs"
	"golang.org/x/mod/modfile"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func TestLogger(t *testing.T) {
	a := assert.New(t, false)

	buf := new(bytes.Buffer)
	ll, err := logs.New(&logs.Options{
		Levels:  logs.AllLevels(),
		Handler: logs.NewTextHandler("", buf),
	})
	a.NotError(err).NotNil(ll)
	l := New(ll, message.NewPrinter(language.SimplifiedChinese))
	a.NotNil(l).Zero(l.Count())

	e1 := &scanner.Error{Pos: token.Position{Filename: "f1.go"}, Msg: "e1"}
	e2 := &scanner.Error{Pos: token.Position{Filename: "f1.go"}, Msg: "e2"}
	l.Error(e1, "f1.go", 0)
	a.Equal(1, l.Count()).True(buf.Len() > 0)

	list := scanner.ErrorList{e1, e2}
	l.Error(list, "f1.go", 0)
	a.Equal(3, l.Count()).True(buf.Len() > 0)

	me := &modfile.Error{
		Err: errors.New("err"),
		Pos: modfile.Position{
			Line:     5,
			LineRune: 10,
		},
	}
	l.Error(me, "f1.go", 0)
	a.Equal(4, l.Count()).True(buf.Len() > 0)

	l.Error(me.Err, "f1.go", 0)
	a.Equal(5, l.Count()).True(buf.Len() > 0)
}

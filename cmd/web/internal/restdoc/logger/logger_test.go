// SPDX-License-Identifier: MIT

package logger

import (
	"bytes"
	"errors"
	"fmt"
	"go/scanner"
	"go/token"
	"testing"

	"github.com/issue9/assert/v3"
	"golang.org/x/mod/modfile"
)

func TestLogger(t *testing.T) {
	a := assert.New(t, false)

	buf := new(bytes.Buffer)
	l := New(func(e *Entry) {
		fmt.Fprintln(buf, e)
	})
	a.NotNil(l).Zero(l.Count())

	e1 := &scanner.Error{Pos: token.Position{Filename: "f1.go"}, Msg: "e1"}
	e2 := &scanner.Error{Pos: token.Position{Filename: "f1.go"}, Msg: "e2"}
	l.LogError(Unknown, e1, "f1.go", 0)
	a.Equal(1, l.Count()).True(buf.Len() > 0)

	list := scanner.ErrorList{e1, e2}
	l.LogError(Unknown, list, "f1.go", 0)
	a.Equal(3, l.Count()).True(buf.Len() > 0)

	me := &modfile.Error{
		Err: errors.New("err"),
		Pos: modfile.Position{
			Line:     5,
			LineRune: 10,
		},
	}
	l.LogError(ModSyntax, me, "f1.go", 0)
	a.Equal(4, l.Count()).True(buf.Len() > 0)

	l.LogError(Unknown, me.Err, "f1.go", 0)
	a.Equal(5, l.Count()).True(buf.Len() > 0)
}

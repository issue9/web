// SPDX-License-Identifier: MIT

package schema

import (
	"fmt"
	"go/token"

	"github.com/issue9/web/cmd/web/restdoc/logger"
)

type Error struct {
	Msg any
	Pos token.Pos
}

func newError(pos token.Pos, msg any) *Error { return &Error{Msg: msg, Pos: pos} }

func (err *Error) Log(l *logger.Logger, fset *token.FileSet) {
	p := fset.Position(err.Pos)
	l.Error(err.Msg, p.Filename, p.Line)
}

func (err *Error) Error() string { return fmt.Sprint(err.Msg) }

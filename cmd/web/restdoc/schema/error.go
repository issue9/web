// SPDX-FileCopyrightText: 2018-2024 caixw
//
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

func (err *Error) Log(l *logger.Logger, fset *token.FileSet) {
	p := fset.Position(err.Pos)
	l.Error(err.Msg, p.Filename, p.Line)
}

func (err *Error) Error() string { return fmt.Sprint(err.Msg) }

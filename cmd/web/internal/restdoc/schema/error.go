// SPDX-License-Identifier: MIT

package schema

import (
	"fmt"
	"go/token"

	"github.com/issue9/web/cmd/web/internal/restdoc/logger"
)

type Error struct {
	Type logger.Type
	Msg  any
	Pos  token.Pos
}

func newError(t logger.Type, pos token.Pos, msg any) *Error {
	return &Error{Type: t, Msg: msg, Pos: pos}
}

func (err *Error) Error() string { return fmt.Sprint(err.Msg) }

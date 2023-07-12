// SPDX-License-Identifier: MIT

package loggertest

import (
	"errors"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/web/cmd/web/internal/restdoc/logger"
)

func TestTester(t *testing.T) {
	a := assert.New(t, false)

	lt := New()
	a.NotNil(lt)
	lt.Log(logger.Cancelled, "aaa", "", 0)
	lt.LogError(logger.GoSyntax, errors.New("text string"), "", 0)

	a.Length(lt.Entries[logger.Cancelled], 1).
		Length(lt.Entries[logger.GoSyntax], 1).
		Length(lt.Entries[logger.DocSyntax], 0)
}

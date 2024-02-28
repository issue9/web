// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package loggertest

import (
	"errors"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/logs/v7"
	"github.com/issue9/web"
)

func TestTester(t *testing.T) {
	a := assert.New(t, false)

	lt := New(a)
	a.NotNil(lt)
	lt.Warning(web.Phrase("aaa"))
	lt.Error(errors.New("text string"), "", 0)

	a.Length(lt.Records[logs.LevelWarn], 1).
		Length(lt.Records[logs.LevelError], 1).
		Length(lt.Records[logs.LevelInfo], 0)
}

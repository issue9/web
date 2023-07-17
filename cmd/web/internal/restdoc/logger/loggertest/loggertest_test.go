// SPDX-License-Identifier: MIT

package loggertest

import (
	"errors"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/web"

	"github.com/issue9/web/logs"
)

func TestTester(t *testing.T) {
	a := assert.New(t, false)

	lt := New(a)
	a.NotNil(lt)
	lt.Warning(web.Phrase("aaa"))
	lt.Error(errors.New("text string"), "", 0)

	a.Length(lt.Records[logs.Warn], 1).
		Length(lt.Records[logs.Error], 1).
		Length(lt.Records[logs.Info], 0)
}

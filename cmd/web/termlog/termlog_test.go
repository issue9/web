// SPDX-License-Identifier: MIT

package termlog

import (
	"os"
	"testing"

	"github.com/issue9/assert/v3"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func TestNew(t *testing.T) {
	a := assert.New(t, false)
	a.NotNil(New(nil, os.Stdout))
	a.NotNil(New(message.NewPrinter(language.SimplifiedChinese), os.Stdout))
}

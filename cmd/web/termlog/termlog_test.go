// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package termlog

import (
	"os"
	"testing"

	"github.com/issue9/assert/v4"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func TestNew(t *testing.T) {
	a := assert.New(t, false)
	a.NotNil(New(nil, os.Stdout)).
		NotNil(New(message.NewPrinter(language.SimplifiedChinese), os.Stdout))
}

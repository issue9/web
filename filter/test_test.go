// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package filter

import (
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/localeutil"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func TestTest(t *testing.T) {
	a := assert.New(t, false)
	p := message.NewPrinter(language.SimplifiedChinese)

	a.Nil(Test(false, nil))

	b1 := NewBuilder(S(trimRight), S(upper))
	v1 := "v1 "
	b2 := NewBuilder(S(upper), V[string](required, localeutil.Phrase("required")))
	v2 := "v2"
	a.Length(Test(true, p, b1("v1", &v1), b2("v2", &v2)), 0).
		Equal(v1, "V1").
		Equal(v2, "V2")

	b2 = NewBuilder(V[string](zero, localeutil.Phrase("zero")))
	a.Equal(Test(true, p, b1("v1", &v1), b2("v2", &v2)), map[string]string{"v2": "zero"}).
		Equal(v1, "V1").
		Equal(v2, "V2")
}

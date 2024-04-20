// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package errs

import (
	"errors"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/localeutil"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
)

func TestSprint(t *testing.T) {
	a := assert.New(t, false)

	a.Panic(func() {
		Sprint(nil, nil, false)
	})

	c := catalog.NewBuilder(catalog.Fallback(language.SimplifiedChinese))
	a.NotError(c.SetString(language.TraditionalChinese, "lang", "tw"))
	a.NotError(c.SetString(language.SimplifiedChinese, "lang", "cn"))
	p := message.NewPrinter(language.SimplifiedChinese, message.Catalog(c))

	a.Equal(Sprint(p, errors.New("lang"), false), "lang")
	a.Equal(Sprint(p, localeutil.Error("lang"), false), "cn")

	a.Contains(Sprint(p, NewDepthStackError(2, errors.New("lang")), true), "lang")
	a.Contains(Sprint(p, NewDepthStackError(2, localeutil.Error("lang")), false), "cn")

	a.Contains(Sprint(p, NewDepthStackError(2, NewDepthStackError(2, errors.New("lang"))), false), "lang")
	a.Contains(Sprint(p, NewDepthStackError(2, NewDepthStackError(2, localeutil.Error("lang"))), true), "cn")
}

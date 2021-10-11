// SPDX-License-Identifier: MIT

package errs

import (
	"errors"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/localeutil"
	"golang.org/x/text/language"
	"golang.org/x/text/message/catalog"
	"gopkg.in/yaml.v2"

	"github.com/issue9/web/locales"
	"github.com/issue9/web/serialization"
)

func TestMerge(t *testing.T) {
	a := assert.New(t)

	err1 := errors.New("err1")
	err2 := errors.New("err2")
	err3 := Merge(err1, err2)
	a.ErrorIs(err3, err1)

	err4 := Merge(err1, nil)
	a.ErrorIs(err4, err1)

	err5 := Merge(nil, err1)
	a.ErrorIs(err5, err1)

	err6 := Merge(nil, nil)
	a.Nil(err6)
}

func TestMergeErrors_LocaleString(t *testing.T) {
	a := assert.New(t)
	hans := language.MustParse("cmn-hans")
	hant := language.MustParse("cmn-hant")

	locale := serialization.NewLocale(catalog.NewBuilder(), serialization.NewFiles(5))
	a.NotError(locale)
	a.NotError(locale.Files().Add(yaml.Marshal, yaml.Unmarshal, ".yaml", ".yml"))
	a.NotError(locale.LoadFileFS(locales.Locales, "*.yml"))

	b := locale.Builder()
	a.NotError(b.SetString(hans, "k1", "cn1"))
	a.NotError(b.SetString(hant, "k1", "tw1"))

	cnp := locale.Printer(hans)
	twp := locale.Printer(hant)

	err := Merge(localeutil.Error("k1"), errors.New("err2"))
	ls, ok := err.(localeutil.LocaleStringer)
	a.True(ok).NotNil(ls)
	a.Equal("在返回 cn1 时再次发生了错误 err2", ls.LocaleString(cnp))
	a.Equal("在返回 tw1 时再次发生了错误 err2", ls.LocaleString(twp))
	a.Equal("err2 when return k1", err.Error())
}

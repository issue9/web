// SPDX-License-Identifier: MIT

package config

import (
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/localeutil"
	"golang.org/x/text/language"
	"golang.org/x/text/message/catalog"
	"gopkg.in/yaml.v2"

	"github.com/issue9/web/locales"
	"github.com/issue9/web/serialization"
)

var (
	_ error                     = &Error{}
	_ localeutil.LocaleStringer = &Error{}
)

func TestError_LocaleString(t *testing.T) {
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

	err := &Error{Message: localeutil.Error("k1"), Config: "path"}
	a.Equal("位于 path: 发生了 cn1", err.LocaleString(cnp))
	a.Equal("位于 path: 发生了 tw1", err.LocaleString(twp))
	a.Equal("k1 at path:", err.LocaleString(localeutil.EmptyPrinter()))
}

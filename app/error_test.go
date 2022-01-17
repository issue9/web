// SPDX-License-Identifier: MIT

package app

import (
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/localeutil"
	"golang.org/x/text/language"
	"golang.org/x/text/message/catalog"
	"gopkg.in/yaml.v2"

	"github.com/issue9/web/locales"
	"github.com/issue9/web/serialization"
)

var (
	_ error                     = &ConfigError{}
	_ localeutil.LocaleStringer = &ConfigError{}
)

func TestError_LocaleString(t *testing.T) {
	a := assert.New(t, false)
	hans := language.MustParse("cmn-hans")
	hant := language.MustParse("cmn-hant")

	locale := serialization.NewLocale(catalog.NewBuilder(), serialization.NewFiles(5))
	a.NotNil(locale)
	a.NotError(locale.Files().Add(yaml.Marshal, yaml.Unmarshal, ".yaml", ".yml"))
	a.NotError(locale.LoadFileFS(locales.Locales, "*.yml"))

	b := locale.Builder()
	a.NotError(b.SetString(hans, "k1", "cn1"))
	a.NotError(b.SetString(hant, "k1", "tw1"))

	cnp := locale.NewPrinter(hans)
	twp := locale.NewPrinter(hant)

	err := &ConfigError{Message: localeutil.Error("k1"), Path: "path"}
	a.Equal("位于 path: 发生了 cn1", err.LocaleString(cnp))
	a.Equal("位于 path: 发生了 tw1", err.LocaleString(twp))
	a.Equal("k1 at path:", err.LocaleString(localeutil.EmptyPrinter()))
}

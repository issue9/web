// SPDX-License-Identifier: MIT

package app

import (
	"testing"
	"time"

	"github.com/issue9/assert/v3"
	"github.com/issue9/localeutil"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"

	"github.com/issue9/web/internal/locale"
	"github.com/issue9/web/internal/serialization"
	"github.com/issue9/web/locales"
)

var (
	_ error                     = &ConfigError{}
	_ localeutil.LocaleStringer = &ConfigError{}
)

func TestError_LocaleString(t *testing.T) {
	a := assert.New(t, false)
	hans := language.MustParse("cmn-hans")
	hant := language.MustParse("cmn-hant")

	f := serialization.NewFS(5)
	l := locale.New(time.UTC, language.MustParse("cmn-hans"))
	a.NotError(f.Serializer().Add(yaml.Marshal, yaml.Unmarshal, ".yaml", ".yml"))
	a.NotError(l.LoadLocaleFiles(locales.Locales, "*.yml", f))

	a.NotError(l.Catalog.SetString(hans, "k1", "cn1"))
	a.NotError(l.Catalog.SetString(hant, "k1", "tw1"))

	cnp := l.NewPrinter(hans)
	twp := l.NewPrinter(hant)

	err := &ConfigError{Message: localeutil.Phrase("k1"), Path: "path"}
	a.Equal("位于 path: 发生了 cn1", err.LocaleString(cnp))
	a.Equal("位于 path: 发生了 tw1", err.LocaleString(twp))
	a.Equal("k1 at path:", err.Error())
}

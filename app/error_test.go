// SPDX-License-Identifier: MIT

package app

import (
	"testing"
	"time"

	"github.com/issue9/assert/v3"
	"github.com/issue9/localeutil"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"

	"github.com/issue9/web/locales"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"
)

var (
	_ error                     = &ConfigError{}
	_ localeutil.LocaleStringer = &ConfigError{}
)

func TestError_LocaleString(t *testing.T) {
	a := assert.New(t, false)
	hans := language.MustParse("cmn-hans")
	hant := language.MustParse("cmn-hant")
	s := servertest.NewServer(a, &server.Options{LanguageTag: language.MustParse("cmn-hans"), Location: time.UTC})
	f := s.Files()
	f.Add(yaml.Marshal, yaml.Unmarshal, ".yaml", ".yml")

	a.NotError(s.LoadLocales(locales.Locales, "*.yml"))

	a.NotError(s.CatalogBuilder().SetString(hans, "k1", "cn1"))
	a.NotError(s.CatalogBuilder().SetString(hant, "k1", "tw1"))

	cnp := s.NewPrinter(hans)
	twp := s.NewPrinter(hant)

	err := &ConfigError{Message: localeutil.Phrase("k1"), Path: "path"}
	a.Equal("位于 path: 发生了 cn1", err.LocaleString(cnp))
	a.Equal("位于 path: 发生了 tw1", err.LocaleString(twp))
	a.Equal("k1 at path:", err.Error())
}

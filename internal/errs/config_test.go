// SPDX-License-Identifier: MIT

package errs_test

import (
	"testing"
	"time"

	"github.com/issue9/assert/v3"
	"github.com/issue9/localeutil"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"

	"github.com/issue9/web/internal/errs"
	"github.com/issue9/web/locales"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"
)

var (
	_ error                     = &errs.ConfigError{}
	_ localeutil.LocaleStringer = &errs.ConfigError{}
)

func TestNewConfigError(t *testing.T) {
	a := assert.New(t, false)

	err1 := errs.NewConfigError("f1", "err1")
	a.NotNil(err1)

	err2 := errs.NewConfigError("f2", err1)
	a.NotNil(err2).
		Equal(err2.Field, "f2.f1").
		Equal(err1.Field, "f2.f1")
}

func TestConfigError_LocaleString(t *testing.T) {
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

	err := errs.NewConfigError("", localeutil.Phrase("k1"))
	err.Path = "path"
	a.Equal("位于 path: 发生了 cn1", err.LocaleString(cnp))
	a.Equal("位于 path: 发生了 tw1", err.LocaleString(twp))
	a.Equal("k1 at path:", err.Error())
}

func TestConfigError_SetFieldParent(t *testing.T) {
	a := assert.New(t, false)

	err := errs.NewConfigError("f1", "error")
	err.AddFieldParent("f2")
	a.Equal(err.Field, "f2.f1")
	err.AddFieldParent("f3")
	a.Equal(err.Field, "f3.f2.f1")
	err.AddFieldParent("")
	a.Equal(err.Field, "f3.f2.f1")

	err = errs.NewConfigError("", "error")
	err.AddFieldParent("f2")
	a.Equal(err.Field, "f2")
	err.AddFieldParent("f3")
	a.Equal(err.Field, "f3.f2")
}

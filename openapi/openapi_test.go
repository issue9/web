// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/web"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
)

var _ web.HandlerFunc = (&Document{}).Handler

func newPrinter(a *assert.Assertion, t language.Tag) *message.Printer {
	b := catalog.NewBuilder()
	a.NotError(b.SetString(language.SimplifiedChinese, "lang", "简体"))
	a.NotError(b.SetString(language.TraditionalChinese, "lang", "繁体"))
	return message.NewPrinter(t, message.Catalog(b))
}

func TestNewLicense(t *testing.T) {
	a := assert.New(t, false)

	a.Equal(newLicense("name", "id"), &licenseRenderer{Name: "name", Identifier: "id"}).
		Equal(newLicense("name", "https://example.com"), &licenseRenderer{Name: "name", URL: "https://example.com"})
}

func TestDocument_build(t *testing.T) {
	a := assert.New(t, false)
	p := newPrinter(a, language.SimplifiedChinese)

	d := New("0.1.0", web.Phrase("lang"))
	r := d.build(p)
	a.Equal(r.Info.Version, "0.1.0").
		Equal(r.OpenAPI, Version).
		Equal(r.Info.Title, "简体")

	d.addOperation("GET", "/users/{id}", "", &Operation{
		Paths: []*Parameter{{Name: "id", Description: web.Phrase("desc")}},
	})
	r = d.build(p)
	a.Equal(r.Info.Version, "0.1.0").
		Equal(r.OpenAPI, Version).
		Equal(r.Info.Title, "简体").
		Equal(r.Paths.Len(), 1)

	d.addOperation("POST", "/users/{id}", "", &Operation{
		Paths: []*Parameter{{Name: "id", Description: web.Phrase("desc")}},
	})
	r = d.build(p)
	obj := r.Paths.GetPair("/users/{id}").Value.obj
	a.NotNil(obj.Get).
		NotNil(obj.Post).
		Nil(obj.Delete)
}

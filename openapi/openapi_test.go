// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"net/http"
	"testing"

	"github.com/issue9/assert/v4"
	"golang.org/x/text/language"

	"github.com/issue9/web"
	"github.com/issue9/web/locales"
	"github.com/issue9/web/server"
)

func newServer(a *assert.Assertion) web.Server {
	s, err := server.NewHTTP("test", "1.0.0", &server.Options{
		HTTPServer: &http.Server{Addr: ":8080"},
	})
	a.NotError(err).NotNil(s)

	s.Locale().LoadMessages("*.yaml", locales.Locales...)
	a.NotError(s.Locale().SetString(language.SimplifiedChinese, "lang", "简体"))
	a.NotError(s.Locale().SetString(language.TraditionalChinese, "lang", "繁体"))

	return s
}

func TestDocument_AddWebHook(t *testing.T) {
	a := assert.New(t, false)
	ss := newServer(a)
	d := New(ss, web.Phrase("desc"))

	d.AddWebhook("hook1", http.MethodGet, &Operation{})
	a.Length(d.webHooks, 1)

	a.PanicString(func() {
		d.AddWebhook("hook1", http.MethodGet, &Operation{})
	}, "已经存在 hook1:GET 的 webhook")
	a.Length(d.webHooks, 1)

	d.AddWebhook("hook1", http.MethodPost, &Operation{})
	a.Length(d.webHooks, 1)
}

func TestParameterizedDescription(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a)
	p := s.Locale().NewPrinter(language.SimplifiedChinese)

	t.Run("panic", func(*testing.T) {
		d := New(s, web.Phrase("title"))

		d.addOperation("GET", "/users/1", "", &Operation{
			d:         d,
			Responses: map[string]*Response{"200": {Body: &Schema{Type: TypeNumber}}},
			ID:        "id1",
		})
		d.AppendDescriptionParameter("id1", "item1", "item2")
		a.PanicString(func() {
			d.build(p, language.SimplifiedChinese, nil)
		}, "接口 id1 未指定 Description 内容").
			Length(d.parameterizedDesc, 1)
	})

	t.Run("not parameterizedDescription", func(*testing.T) {
		d := New(s, web.Phrase("title"))

		d.addOperation("GET", "/users/2", "", &Operation{
			d:           d,
			Responses:   map[string]*Response{"200": {Body: &Schema{Type: TypeNumber}}},
			ID:          "id2",
			Description: web.Phrase("desc"),
		})

		d.AppendDescriptionParameter("id2", "item1", "item2")
		r := d.build(p, language.SimplifiedChinese, nil)
		a.Equal(r.Paths.GetPair("/users/2").Value.obj.Get.Description, "desc")
	})

	t.Run("parameterizedDescription", func(*testing.T) {
		d := New(s, web.Phrase("title"))

		d.addOperation("GET", "/users/3", "", &Operation{
			d:           d,
			Responses:   map[string]*Response{"200": {Body: &Schema{Type: TypeNumber}}},
			ID:          "id3",
			Description: ParameterizedDescription("desc %s", nil),
		})

		d.AppendDescriptionParameter("id3", "item1", "item2")
		r := d.build(p, language.SimplifiedChinese, nil)
		a.Equal(r.Paths.GetPair("/users/3").Value.obj.Get.Description, "desc item1\nitem2")
	})
}

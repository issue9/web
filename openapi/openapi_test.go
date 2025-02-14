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

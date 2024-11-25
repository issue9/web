// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package html_test

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/issue9/assert/v4"
	"github.com/issue9/logs/v7"
	"github.com/issue9/mux/v9/header"
	"golang.org/x/text/language"

	"github.com/issue9/web"
	"github.com/issue9/web/mimetype/html"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"
)

func newServer(a *assert.Assertion, lang string) web.Server {
	s, err := server.NewHTTP("test", "1.0.0", &server.Options{
		HTTPServer: &http.Server{Addr: ":8080"},
		Language:   language.MustParse(lang),
		Codec:      web.NewCodec().AddMimetype(html.Mimetype, html.Marshal, html.Unmarshal, ""),
		Logs:       logs.New(logs.NewTermHandler(os.Stderr, nil)),
	})
	a.NotError(err).NotNil(s)

	// locale
	b := s.Locale()
	a.NotError(b.SetString(language.Und, "lang", "und"))
	a.NotError(b.SetString(language.MustParse("cmn-hans"), "lang", "hans"))
	a.NotError(b.SetString(language.MustParse("cmn-hant"), "lang", "hant"))

	return s
}

func TestInstall(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a, "und") // und 被解析为 language.Und，将会尝试读取系统的本地化信息
	html.Install(s, nil, nil, "*.tpl", os.DirFS("./testdata/view"))

	defer servertest.Run(a, s)()
	defer s.Close(500 * time.Millisecond)

	type obj struct {
		XMLName struct{} `html:"t"`
	}

	r := s.Routers().New("def", nil)
	r.Get("/path", func(ctx *web.Context) web.Responser {
		return web.ResponserFunc(func(ctx *web.Context) { ctx.Render(200, &obj{}) })
	})

	servertest.Get(a, "http://localhost:8080/path").
		Header(header.AcceptLanguage, "cmn-hans").
		Header(header.Accept, html.Mimetype).
		Do(nil).
		Status(200).
		StringBody("\n<div>hans</div>\n<div>hans</div>\n")

	servertest.Get(a, "http://localhost:8080/path").
		Header(header.AcceptLanguage, "zh-hant").
		Header(header.Accept, html.Mimetype).
		Do(nil).
		Status(200).
		StringBody("\n<div>hant</div>\n<div>hans</div>\n")
}

func TestInstall_withLocalized(t *testing.T) {
	localized := map[language.Tag]string{language.SimplifiedChinese: "cmn-hans", language.TraditionalChinese: "cmn-hant", language.Und: "cmn-hans"}
	a := assert.New(t, false)
	s := newServer(a, "cmn-hans")
	html.Install(s, nil, localized, "*.tpl", os.DirFS("./testdata/dir"))

	defer servertest.Run(a, s)()
	defer s.Close(500 * time.Millisecond)

	type obj struct {
		XMLName struct{} `html:"t"`
	}

	r := s.Routers().New("def", nil)
	r.Get("/path", func(ctx *web.Context) web.Responser {
		return web.ResponserFunc(func(ctx *web.Context) { ctx.Render(200, &obj{}) })
	})
	servertest.Get(a, "http://localhost:8080/path").
		Header(header.AcceptLanguage, "cmn-hans").
		Header(header.Accept, html.Mimetype).
		Do(nil).
		Status(200).
		StringBody("\n<div>hans简</div>\n<div>hans</div>\n")

	servertest.Get(a, "http://localhost:8080/path").
		Header(header.AcceptLanguage, "cmn-hant").
		Header(header.Accept, html.Mimetype).
		Do(nil).
		Status(200).
		StringBody("\n<div>hant繁</div>\n<div>hans</div>\n")
}

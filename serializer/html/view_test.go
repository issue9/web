// SPDX-License-Identifier: MIT

package html

import (
	"net/http"
	"os"
	"testing"

	"github.com/issue9/assert/v3"
	"golang.org/x/text/language"

	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"
)

func newServer(a *assert.Assertion, lang string) *server.Server {
	s, err := server.New("test", "1.0.0", &server.Options{
		HTTPServer: &http.Server{Addr: ":8080"},
		Locale:     &server.Locale{Language: language.MustParse(lang)},
		Mimetypes: []*server.Mimetype{
			{
				Type:      Mimetype,
				Marshal:   Marshal,
				Unmarshal: Unmarshal,
			},
		},
	})
	a.NotError(err).NotNil(s)

	// locale
	b := s.CatalogBuilder()
	a.NotError(b.SetString(language.Und, "lang", "und"))
	a.NotError(b.SetString(language.SimplifiedChinese, "lang", "hans"))
	a.NotError(b.SetString(language.TraditionalChinese, "lang", "hant"))

	return s
}

func TestInstallView(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a, "und")
	InstallView(s, false, os.DirFS("./testdata/view"), "*.tpl")

	defer servertest.Run(a, s)()
	defer s.Close(0)

	type obj struct {
		HTMLName struct{} `html:"t"`
	}

	r := s.Routers().New("def", nil)
	r.Get("/path", func(ctx *server.Context) server.Responser {
		return server.ResponserFunc(func(ctx *server.Context) {
			ctx.Render(200, &obj{}, false)
		})
	})

	servertest.Get(a, "http://localhost:8080/path").
		Header("accept-language", "cmn-hans").
		Header("accept", Mimetype).
		Do(nil).
		Status(200).
		StringBody("\n<div>hans</div>\n<div>hans</div>\n")

	servertest.Get(a, "http://localhost:8080/path").
		Header("accept-language", "zh-hant").
		Header("accept", Mimetype).
		Do(nil).
		Status(200).
		StringBody("\n<div>hant</div>\n<div>hans</div>\n")
}

func TestInstallView_dir(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a, "cmn-hans")
	instalDirView(s, os.DirFS("./testdata/dir"), "*.tpl")

	defer servertest.Run(a, s)()
	defer s.Close(0)

	type obj struct {
		HTMLName struct{} `html:"t"`
	}

	r := s.Routers().New("def", nil)
	r.Get("/path", func(ctx *server.Context) server.Responser {
		return server.ResponserFunc(func(ctx *server.Context) {
			ctx.Render(200, &obj{}, false)
		})
	})
	servertest.Get(a, "http://localhost:8080/path").
		Header("accept-language", "cmn-hans").
		Header("accept", Mimetype).
		Do(nil).
		Status(200).
		StringBody("\n<div>hans简</div>\n<div>hans</div>\n")

	servertest.Get(a, "http://localhost:8080/path").
		Header("accept-language", "cmn-hant").
		Header("accept", Mimetype).
		Do(nil).
		Status(200).
		StringBody("\n<div>hant繁</div>\n<div>hans</div>\n")

	// 默认语言
	servertest.Get(a, "http://localhost:8080/path").
		Header("accept-language", "en").
		Header("accept", Mimetype).
		Do(nil).
		Status(200).
		StringBody("\n<div>und简</div>\n<div>hans</div>\n")
}

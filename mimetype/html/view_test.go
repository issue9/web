// SPDX-License-Identifier: MIT

package html

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/issue9/assert/v3"
	"golang.org/x/text/language"

	"github.com/issue9/web"
	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"
)

func newServer(a *assert.Assertion, lang string) web.Server {
	s, err := server.New("test", "1.0.0", &server.Options{
		HTTPServer: &http.Server{Addr: ":8080"},
		Language:   language.MustParse(lang),
		Mimetypes: []*server.Mimetype{
			{
				Name:      Mimetype,
				Marshal:   Marshal,
				Unmarshal: Unmarshal,
			},
		},
	})
	a.NotError(err).NotNil(s)

	// locale
	b := s.Locale()
	a.NotError(b.SetString(language.Und, "lang", "und"))
	a.NotError(b.SetString(language.MustParse("cmn-hans"), "lang", "hans"))
	a.NotError(b.SetString(language.MustParse("cmn-hant"), "lang", "hant"))

	return s
}

func TestInstallView(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a, "und")
	InstallView(s, false, os.DirFS("./testdata/view"), "*.tpl")

	defer servertest.Run(a, s)()
	defer s.Close(500 * time.Millisecond)

	type obj struct {
		HTMLName struct{} `html:"t"`
	}

	r := s.NewRouter("def", nil)
	r.Get("/path", func(ctx *web.Context) web.Responser {
		return web.ResponserFunc(func(ctx *web.Context) *web.Problem {
			ctx.Render(200, &obj{})
			return nil
		})
	})

	servertest.Get(a, "http://localhost:8080/path").
		Header(header.AcceptLang, "cmn-hans").
		Header(header.Accept, Mimetype).
		Do(nil).
		Status(200).
		StringBody("\n<div>hans</div>\n<div>hans</div>\n")

	servertest.Get(a, "http://localhost:8080/path").
		Header(header.AcceptLang, "zh-hant").
		Header(header.Accept, Mimetype).
		Do(nil).
		Status(200).
		StringBody("\n<div>hant</div>\n<div>hans</div>\n")
}

func TestInstallDirView(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a, "cmn-hans")
	instalDirView(s, os.DirFS("./testdata/dir"), "*.tpl")

	defer servertest.Run(a, s)()
	defer s.Close(500 * time.Millisecond)

	type obj struct {
		HTMLName struct{} `html:"t"`
	}

	r := s.NewRouter("def", nil)
	r.Get("/path", func(ctx *web.Context) web.Responser {
		return web.ResponserFunc(func(ctx *web.Context) *web.Problem {
			ctx.Render(200, &obj{})
			return nil
		})
	})
	servertest.Get(a, "http://localhost:8080/path").
		Header(header.AcceptLang, "cmn-hans").
		Header(header.Accept, Mimetype).
		Do(nil).
		Status(200).
		StringBody("\n<div>hans简</div>\n<div>hans</div>\n")

	servertest.Get(a, "http://localhost:8080/path").
		Header(header.AcceptLang, "cmn-hant").
		Header(header.Accept, Mimetype).
		Do(nil).
		Status(200).
		StringBody("\n<div>hant繁</div>\n<div>hans</div>\n")
}

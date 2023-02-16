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

func TestInstallView(t *testing.T) {
	a := assert.New(t, false)
	s := servertest.NewTester(a, nil)
	s.Server().Mimetypes().Add(Mimetype, Marshal, Unmarshal, "")
	InstallView(s.Server(), false, os.DirFS("./testdata/view"), "*.tpl")

	s.GoServe()
	defer s.Close(0)

	type obj struct {
		HTMLName struct{} `html:"t"`
	}

	r := s.Router()
	r.Get("/path", func(ctx *server.Context) server.Responser {
		return server.ResponserFunc(func(ctx *server.Context) {
			ctx.Marshal(200, &obj{}, false)
		})
	})

	s.Get("/path").
		Header("accept-language", "cmn-hans").
		Header("accept", Mimetype).
		Do(nil).
		Status(200).
		StringBody("\n<div>hans</div>\n<div>hans</div>\n")

	s.Get("/path").
		Header("accept-language", "zh-hant").
		Header("accept", Mimetype).
		Do(nil).
		Status(200).
		StringBody("\n<div>hant</div>\n<div>hans</div>\n")
}

func TestInstallView_dir(t *testing.T) {
	a := assert.New(t, false)
	opt := &server.Options{
		HTTPServer: &http.Server{Addr: ":8080"},
		Locale:     &server.Locale{Language: language.MustParse("cmn-hans")},
	}
	s := servertest.NewTester(a, opt)
	s.Server().Mimetypes().Add(Mimetype, Marshal, Unmarshal, "")
	instalDirView(s.Server(), os.DirFS("./testdata/dir"), "*.tpl")

	s.GoServe()
	defer s.Close(0)

	type obj struct {
		HTMLName struct{} `html:"t"`
	}

	r := s.Router()
	r.Get("/path", func(ctx *server.Context) server.Responser {
		return server.ResponserFunc(func(ctx *server.Context) {
			ctx.Marshal(200, &obj{}, false)
		})
	})
	s.Get("/path").
		Header("accept-language", "cmn-hans").
		Header("accept", Mimetype).
		Do(nil).
		Status(200).
		StringBody("\n<div>hans简</div>\n<div>hans</div>\n")

	s.Get("/path").
		Header("accept-language", "cmn-hant").
		Header("accept", Mimetype).
		Do(nil).
		Status(200).
		StringBody("\n<div>hant繁</div>\n<div>hans</div>\n")

	// 默认语言
	s.Get("/path").
		Header("accept-language", "en").
		Header("accept", Mimetype).
		Do(nil).
		Status(200).
		StringBody("\n<div>und简</div>\n<div>hans</div>\n")
}

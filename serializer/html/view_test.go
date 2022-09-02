// SPDX-License-Identifier: MIT

package html

import (
	"os"
	"testing"

	"github.com/issue9/assert/v3"

	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"
)

func TestView(t *testing.T) {
	a := assert.New(t, false)
	s := servertest.NewTester(a, nil)
	s.Server().Mimetypes().Add(Marshal, Unmarshal, Mimetype)
	v := NewView(s.Server(), os.DirFS("./testdata/view"), "*.tpl", false)
	a.NotNil(v)

	s.GoServe()
	defer s.Close(0)

	r := s.NewRouter()
	r.Get("/path", func(ctx *server.Context) server.Responser {
		return server.ResponserFunc(func(ctx *server.Context) {
			ctx.Marshal(200, v.View(ctx, "t", nil), false)
		})
	})
	s.Get("/path").
		Header("accept-language", "cmn-hans").
		Header("accept", Mimetype).
		Do(nil).
		Status(200).
		StringBody("\n<div>hans</div>\n")
}

func TestView_locale(t *testing.T) {
	a := assert.New(t, false)
	s := servertest.NewTester(a, nil)
	s.Server().Mimetypes().Add(Marshal, Unmarshal, Mimetype)
	v := NewView(s.Server(), os.DirFS("./testdata/localeview"), "*.tpl", true)
	a.NotNil(v)

	s.GoServe()
	defer s.Close(0)

	r := s.NewRouter()
	r.Get("/path", func(ctx *server.Context) server.Responser {
		return server.ResponserFunc(func(ctx *server.Context) {
			ctx.Marshal(200, v.View(ctx, "t", nil), false)
		})
	})
	s.Get("/path").
		Header("accept-language", "cmn-hans").
		Header("accept", Mimetype).
		Do(nil).
		Status(200).
		StringBody("\n<div>hans简</div>\n")

	s.Get("/path").
		Header("accept-language", "cmn-hant").
		Header("accept", Mimetype).
		Do(nil).
		Status(200).
		StringBody("\n<div>hant繁</div>\n")
}

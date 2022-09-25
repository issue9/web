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
	s.Server().Mimetypes().Add(Marshal, Unmarshal, Mimetype)
	InstallView(s.Server(), os.DirFS("./testdata/view"), "*.tpl")

	s.GoServe()
	defer s.Close(0)

	type obj struct {
		HTMLName struct{} `html:"t"`
	}

	r := s.NewRouter()
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

func TestNewLocaleView(t *testing.T) {
	a := assert.New(t, false)
	s := servertest.NewTester(a, &server.Options{HTTPServer: &http.Server{Addr: ":8080"}, LanguageTag: language.MustParse("cmn-hans")})
	s.Server().Mimetypes().Add(Marshal, Unmarshal, Mimetype)
	InstallLocaleView(s.Server(), os.DirFS("./testdata/localeview"), "*.tpl")

	s.GoServe()
	defer s.Close(0)

	type obj struct {
		HTMLName struct{} `html:"t"`
	}

	r := s.NewRouter()
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

func TestGetName(t *testing.T) {
	a := assert.New(t, false)

	type obj struct {
		HTMLName struct{} `html:"t"`
	}
	type obj2 struct {
		HTMLName struct{}
	}

	type obj3 struct{}

	type obj4 map[string]string

	a.Equal(getName(&obj{}), "t")
	a.Equal(getName(&obj2{}), "obj2")
	a.Equal(getName(&obj3{}), "obj3")
	a.Equal(getName(&obj4{}), "obj4")

	a.PanicString(func() {
		getName(map[string]string{})
	}, "text/html 不支持输出当前类型的对象")
}

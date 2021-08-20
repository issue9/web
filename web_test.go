// SPDX-License-Identifier: MIT

package web

import (
	"bytes"
	"net/http"
	"os"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/logs/v3"
	"github.com/issue9/mux/v5/group"
	"golang.org/x/text/language"
	"golang.org/x/text/message/catalog"
	"gopkg.in/yaml.v2"

	"github.com/issue9/web/serialization"
)

func TestNewServer(t *testing.T) {
	a := assert.New(t)

	locale := serialization.NewLocale(catalog.NewBuilder(), serialization.NewFiles(5))
	a.NotError(locale)
	a.NotError(locale.Files().Add(yaml.Marshal, yaml.Unmarshal, ".yaml", ".yml"))
	a.NotError(locale.LoadFileFS(os.DirFS("./locales"), "*.yml"))

	buf := new(bytes.Buffer)
	log, err := logs.New(nil)
	a.NotError(err).NotNil(log)
	a.NotError(log.SetOutput(logs.LevelInfo, buf))

	srv, err := NewServer("app", "v1.1", &Options{Locale: locale, Logs: log, Tag: language.MustParse("cmn-hans")})
	a.NotError(err).NotNil(srv)

	m1, err := srv.NewModule("m1", "1.0.0", "m1 desc")
	a.NotError(err).NotNil(m1)
	m1.Tag("init").AddInit("m1 init", func() error {
		router, err := srv.NewRouter("r1", "https://example.com", group.MatcherFunc(group.Any))
		a.NotError(err).NotNil(router)
		return nil
	})
	m1.Tag("init").AddRoutes(func(r *Router) {
		r.Get("/path", func(c *Context) Responser {
			// do something
			return Status(http.StatusCreated)
		})
	}, "r1")

	a.NotError(srv.InitModules("init"))

	a.Contains(buf.String(), "注册路由: r1") // 查看是否正确加载翻译内容
}

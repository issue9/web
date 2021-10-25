// SPDX-License-Identifier: MIT

package web

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/logs/v3"
	"github.com/issue9/mux/v5/group"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"

	"github.com/issue9/web/locales"
	"github.com/issue9/web/serialization"
	"github.com/issue9/web/serialization/text"
	"github.com/issue9/web/serialization/text/testobject"
)

func TestNewServer(t *testing.T) {
	a := assert.New(t)

	files := serialization.NewFiles(5)
	a.NotError(files)
	a.NotError(files.Add(yaml.Marshal, yaml.Unmarshal, ".yaml", ".yml"))

	buf := new(bytes.Buffer)
	log, err := logs.New(nil)
	a.NotError(err).NotNil(log)
	a.NotError(log.SetOutput(logs.LevelInfo, buf))

	srv, err := NewServer("app", "v1.1", &Options{Files: files, Logs: log, Tag: language.MustParse("cmn-hans")})
	a.NotError(err).NotNil(srv)

	a.NotError(srv.Locale().LoadFileFS(locales.Locales, "*.yml")) // 加载本地信息

	m1 := srv.NewModule("m1", "1.0.0", Phrase("m1 desc"))
	a.NotNil(m1)
	m1.Action("init").AddInit("m1 init", func() error {
		router, err := srv.NewRouter("r1", "https://example.com", group.MatcherFunc(group.Any))
		a.NotError(err).NotNil(router)
		return nil
	})
	m1.Action("init").AddRoutes(func(r *Router) {
		r.Get("/path", func(c *Context) Responser {
			// do something
			return Created(nil, "")
		})
	}, "r1")

	a.NotError(srv.Serve(false, "init"))

	a.Contains(buf.String(), "注册路由: r1") // 查看是否正确加载翻译内容
}

func TestCreated(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()
	s, err := NewServer("test", "1.0", nil)
	a.NotError(s.Mimetypes().Add(text.Marshal, text.Unmarshal, text.Mimetype))
	a.NotError(err).NotNil(s)

	r := httptest.NewRequest(http.MethodPost, "/path", nil)
	r.Header.Set("Accept", text.Mimetype)
	r.Header.Set("content-type", text.Mimetype)
	resp := Created(&testobject.TextObject{Name: "test", Age: 123}, "")
	ctx := s.NewContext(w, r)
	a.NotError(ctx.Marshal(resp.Status(), resp.Body(), resp.Headers()))
	a.Equal(w.Code, http.StatusCreated).
		Equal(w.Body.String(), `test,123`)

	w.Body.Reset()
	r = httptest.NewRequest(http.MethodPost, "/path", nil)
	r.Header.Set("Accept", text.Mimetype)
	r.Header.Set("content-type", text.Mimetype)
	resp = Created(&testobject.TextObject{Name: "test", Age: 123}, "/test")
	ctx = s.NewContext(w, r)
	a.NotError(ctx.Marshal(resp.Status(), resp.Body(), resp.Headers()))
	a.Equal(w.Code, http.StatusCreated).
		Equal(w.Body.String(), `test,123`).
		Equal(w.Header().Get("Location"), "/test")
}

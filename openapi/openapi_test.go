// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	stdjson "encoding/json"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/issue9/assert/v4"
	"golang.org/x/text/language"

	"github.com/issue9/web"
	"github.com/issue9/web/mimetype/json"
	"github.com/issue9/web/mimetype/yaml"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"
)

var _ web.HandlerFunc = (&Document{}).Handler

func newServer(a *assert.Assertion) web.Server {
	s, err := server.NewHTTP("test", "1.0.0", &server.Options{
		HTTPServer: &http.Server{Addr: ":8080"},
	})
	a.NotError(err).NotNil(s)

	a.NotError(s.Locale().SetString(language.SimplifiedChinese, "lang", "简体"))
	a.NotError(s.Locale().SetString(language.TraditionalChinese, "lang", "繁体"))

	return s
}

func TestNewLicense(t *testing.T) {
	a := assert.New(t, false)

	a.Equal(newLicense("name", "id"), &licenseRenderer{Name: "name", Identifier: "id"}).
		Equal(newLicense("name", "https://example.com"), &licenseRenderer{Name: "name", URL: "https://example.com"})
}

func TestDocument_build(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a)
	p := s.Locale().NewPrinter(language.SimplifiedChinese)

	d := New(s, web.Phrase("lang"))
	r := d.build(p, nil)
	a.Equal(r.Info.Version, s.Version()).
		Equal(r.OpenAPI, Version).
		Equal(r.Info.Title, "简体")

	d.addOperation("GET", "/users/{id}", "", &Operation{
		Paths:     []*Parameter{{Name: "id", Description: web.Phrase("desc")}},
		Responses: map[string]*Response{"200": {Body: &Schema{Type: TypeNumber}}},
	})
	r = d.build(p, nil)
	a.Equal(r.Info.Version, s.Version()).
		Equal(r.OpenAPI, Version).
		Equal(r.Info.Title, "简体").
		Equal(r.Paths.Len(), 1)

	d.addOperation("POST", "/users/{id}", "", &Operation{
		Tags:      []string{"admin"},
		Paths:     []*Parameter{{Name: "id", Description: web.Phrase("desc")}},
		Responses: map[string]*Response{"200": {Body: &Schema{Type: TypeNumber}}},
	})
	r = d.build(p, nil)
	obj := r.Paths.GetPair("/users/{id}").Value.obj
	a.NotNil(obj.Get).
		NotNil(obj.Post).
		Nil(obj.Delete)

	// 带过滤

	r = d.build(p, []string{"admin"})
	obj = r.Paths.GetPair("/users/{id}").Value.obj
	a.Nil(obj.Get).
		NotNil(obj.Post).
		Nil(obj.Delete)
}

func TestDocument_Handler(t *testing.T) {
	a := assert.New(t, false)
	s, err := server.NewHTTP("test", "1.0.0", &server.Options{
		HTTPServer: &http.Server{Addr: ":8080"},
	})
	a.NotError(err).NotNil(s)

	d := New(s, web.Phrase("test"))
	r := s.Routers().New("def", nil)
	a.NotNil(r)

	r.Prefix("/p").
		Delete("/users", func(ctx *web.Context) web.Responser { return nil }, d.API(func(o *Operation) {
			o.Response("200", 1, web.Phrase("get users"), nil)
		})).
		Get("/users", func(*web.Context) web.Responser { return nil }). // 未指定文档
		Get("/openapi", d.Handler)

	cancel := servertest.Run(a, s)

	servertest.Get(a, "http://localhost:8080/p/openapi").Header("accept", json.Mimetype).
		Do(nil).
		BodyFunc(func(a *assert.Assertion, body []byte) {
			a.NotZero(len(body)).True(stdjson.Valid(body))
		})

	servertest.Get(a, "http://localhost:8080/p/openapi").Header("accept", yaml.Mimetype).
		Do(nil).
		Status(http.StatusNotAcceptable) // server 中未未配置 yaml

	d.Disable(true)
	servertest.Get(a, "http://localhost:8080/p/openapi").Header("accept", json.Mimetype).
		Do(nil).
		Status(http.StatusNotImplemented)

	s.Close(500 * time.Millisecond)
	cancel()
}

func TestComponents_build(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a)
	p := s.Locale().NewPrinter(language.SimplifiedChinese)
	d := New(s, web.Phrase("lang"))
	c := d.components

	c.queries["q1"] = &Parameter{Name: "q1", Schema: &Schema{Type: TypeString}}
	c.cookies["c1"] = &Parameter{Name: "c1", Schema: &Schema{Type: TypeNumber}}
	c.headers["h1"] = &Parameter{Name: "h1", Schema: &Schema{Type: TypeBoolean}}
	c.schemas["s1"] = NewSchema(reflect.TypeFor[int](), nil, nil)

	r := c.build(p, d)
	a.Equal(r.Parameters.Len(), 2).
		Equal(r.Headers.Len(), 1).
		Equal(r.Schemas.Len(), 1)
}

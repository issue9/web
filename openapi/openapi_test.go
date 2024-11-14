// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	stdjson "encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/issue9/assert/v4"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web"
	"github.com/issue9/web/mimetype/json"
	"github.com/issue9/web/mimetype/yaml"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"
)

var _ web.HandlerFunc = (&Document{}).Handler

func newPrinter(a *assert.Assertion, t language.Tag) *message.Printer {
	b := catalog.NewBuilder()
	a.NotError(b.SetString(language.SimplifiedChinese, "lang", "简体"))
	a.NotError(b.SetString(language.TraditionalChinese, "lang", "繁体"))
	return message.NewPrinter(t, message.Catalog(b))
}

func TestNewLicense(t *testing.T) {
	a := assert.New(t, false)

	a.Equal(newLicense("name", "id"), &licenseRenderer{Name: "name", Identifier: "id"}).
		Equal(newLicense("name", "https://example.com"), &licenseRenderer{Name: "name", URL: "https://example.com"})
}

func TestDocument_build(t *testing.T) {
	a := assert.New(t, false)
	p := newPrinter(a, language.SimplifiedChinese)

	d := New("0.1.0", web.Phrase("lang"))
	r := d.build(p)
	a.Equal(r.Info.Version, "0.1.0").
		Equal(r.OpenAPI, Version).
		Equal(r.Info.Title, "简体")

	d.addOperation("GET", "/users/{id}", "", &Operation{
		Paths: []*Parameter{{Name: "id", Description: web.Phrase("desc")}},
	})
	r = d.build(p)
	a.Equal(r.Info.Version, "0.1.0").
		Equal(r.OpenAPI, Version).
		Equal(r.Info.Title, "简体").
		Equal(r.Paths.Len(), 1)

	d.addOperation("POST", "/users/{id}", "", &Operation{
		Paths: []*Parameter{{Name: "id", Description: web.Phrase("desc")}},
	})
	r = d.build(p)
	obj := r.Paths.GetPair("/users/{id}").Value.obj
	a.NotNil(obj.Get).
		NotNil(obj.Post).
		Nil(obj.Delete)
}

func TestDocument_Handler(t *testing.T) {
	a := assert.New(t, false)
	s, err := server.NewHTTP("test", "1.0.0", &server.Options{})
	a.NotError(err).NotNil(s)

	d := New("1.0.0", web.Phrase("test"))
	r := s.Routers().New("def", nil)
	a.NotNil(r)
	r.Delete("/users", func(ctx *web.Context) web.Responser { return nil }, d.API().Response(200, 1, web.Phrase("get users"))).
		Get("/users", func(*web.Context) web.Responser { return nil }). // 未指定文档
		Get("/openapi", d.Handler)
	cancel := servertest.Run(a, s)
	servertest.Get(a, "http://localhost:8080/openapi").Header("accept", json.Mimetype).
		Do(nil).
		BodyFunc(func(a *assert.Assertion, body []byte) {
			a.NotZero(len(body)).True(stdjson.Valid(body))
		})
	servertest.Get(a, "http://localhost:8080/openapi").Header("accept", yaml.Mimetype).
		Do(nil).
		Status(http.StatusNotAcceptable) // server 中未未配置 yaml
	s.Close(500 * time.Millisecond)
	cancel()
}

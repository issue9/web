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
	stdyaml "gopkg.in/yaml.v3"

	"github.com/issue9/web"
	"github.com/issue9/web/mimetype/json"
	"github.com/issue9/web/mimetype/yaml"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"
)

var (
	_ stdjson.Marshaler = &renderer[int]{}
	_ stdyaml.Marshaler = &renderer[int]{}
)

type object struct { // 被用于多种用途，所以同时带了 XML 和 yaml。
	XMLName struct{}  `json:"-" yaml:"-" xml:"object"`
	ID      int       `json:"id" xml:"Id" yaml:"id,omitempty"`
	Items   []*object // 循环引用
	Name    string    `json:"name,omitempty" xml:"name,omitempty" yaml:"name,omitempty"`
}

func TestRenderer(t *testing.T) {
	a := assert.New(t, false)
	ss := newServer(a)
	p := ss.Locale().NewPrinter(language.SimplifiedChinese)

	a.PanicString(func() {
		newRenderer[int](nil, nil)
	}, "ref 和 obj 不能同时为 nil")

	ref := &Ref{Ref: "ref"}

	r := newRenderer[object](ref.build(p, "schemas"), &object{ID: 2})
	a.Equal(r.ref.Ref, "#/components/schemas/ref").
		Empty(r.ref.Summary).
		NotNil(r.obj)

	// JSON
	bs, err := stdjson.Marshal(r)
	a.NotError(err).Equal(string(bs), `{"$ref":"#/components/schemas/ref"}`)

	// YAML
	bs, err = stdyaml.Marshal(r)
	a.NotError(err).Equal(string(bs), "$ref: '#/components/schemas/ref'\n")

	// ref = nil

	r = newRenderer(nil, &object{ID: 2})
	a.Nil(r.ref).NotNil(r.obj)

	// JSON
	bs, err = stdjson.Marshal(r)
	a.NotError(err).Equal(string(bs), `{"id":2,"Items":null}`)

	// YAML
	bs, err = stdyaml.Marshal(r)
	a.NotError(err).Equal(string(bs), "id: 2\nitems: []\n")
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
		Get("/openapi", d.Handler())

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

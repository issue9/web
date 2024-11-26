// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"net/http"
	"testing"

	"github.com/issue9/assert/v4"

	"github.com/issue9/web"
)

type q struct {
	Q1 string `query:"q1"`
	Q2 int    `query:"q2"`
	Q3 int
}

func newOperation(a *assert.Assertion) *Operation {
	ss := newServer(a)
	d := New(ss, web.Phrase("title"))
	o := &Operation{
		d:         d,
		Responses: make(map[string]*Response, len(d.responses)+1), // 必然存在的字段，直接初始化了。
	}
	return o
}

func TestOperation_Server(t *testing.T) {
	a := assert.New(t, false)
	o := newOperation(a)

	o.Server("https://example.com", web.Phrase("lang"))
	a.Length(o.Servers, 1)

	a.Panic(func() {
		o.Server("https://example.com/{id}", web.Phrase("lang"))
	})
}

func TestOperation_Path(t *testing.T) {
	a := assert.New(t, false)
	o := newOperation(a)

	o.PathID("p1", nil)
	a.Length(o.Paths, 1)

	o.Path("p2", TypeInteger, nil, nil)
	a.Length(o.Paths, 2)

	a.PanicString(func() {
		o.PathRef("p3", nil, nil)
	}, "未找到引用 p3")

	o.d.components.paths["p3"] = &Parameter{Schema: &Schema{Type: TypeInteger}}
	o.PathRef("p3", nil, nil)
	a.Length(o.Paths, 3)
}

func TestOperation_Query(t *testing.T) {
	a := assert.New(t, false)
	o := newOperation(a)

	o.Query("q1", TypeInteger, nil, nil)
	a.Length(o.Queries, 1)

	a.PanicString(func() {
		o.QueryRef("q2", nil, nil)
	}, "未找到引用 q2")

	o.d.components.queries["q2"] = &Parameter{Schema: &Schema{Type: TypeInteger}}
	o.QueryRef("q2", nil, nil)
	a.Length(o.Queries, 2)
}

func TestOperation_Cookie(t *testing.T) {
	a := assert.New(t, false)
	o := newOperation(a)

	o.Cookie("c1", TypeInteger, nil, nil)
	a.Length(o.Cookies, 1)

	a.PanicString(func() {
		o.CookieRef("c2", nil, nil)
	}, "未找到引用 c2")

	o.d.components.cookies["c2"] = &Parameter{Schema: &Schema{Type: TypeInteger}}
	o.CookieRef("c2", nil, nil)
	a.Length(o.Cookies, 2)
}

func TestOperation_Header(t *testing.T) {
	a := assert.New(t, false)
	o := newOperation(a)

	o.Header("h1", TypeInteger, nil, nil)
	a.Length(o.Headers, 1)

	a.PanicString(func() {
		o.HeaderRef("h2", nil, nil)
	}, "未找到引用 h2")

	o.d.components.headers["h2"] = &Parameter{Schema: &Schema{Type: TypeInteger}}
	o.HeaderRef("h2", nil, nil)
	a.Length(o.Headers, 2)
}

func TestOperation_Response(t *testing.T) {
	a := assert.New(t, false)
	o := newOperation(a)

	o.Response("2xx", object{}, nil, nil)
	a.Length(o.Responses, 1)

	a.PanicString(func() {
		o.ResponseRef("301", "3xx", nil, nil)
	}, "未找到引用 3xx")

	o.d.components.responses["3xx"] = &Response{Body: &Schema{Type: TypeInteger}}
	o.ResponseRef("301", "3xx", nil, nil)
	a.Length(o.Responses, 2)
}

func TestOperation_Callback(t *testing.T) {
	a := assert.New(t, false)
	o := newOperation(a)

	o.Callback("c1", "/path1", http.MethodGet, nil)
	a.Length(o.Callbacks, 1)

	a.PanicString(func() {
		o.CallbackRef("c2", "c2", nil, nil)
	}, "未找到引用 c2")

	o.d.components.callbacks["c2"] = &Callback{}
	o.CallbackRef("c2", "c2", nil, nil)
	a.Length(o.Callbacks, 2)
}

func TestDocument_API(t *testing.T) {
	a := assert.New(t, false)
	ss := newServer(a)
	d := New(ss, web.Phrase("title"))

	m := d.API(func(o *Operation) {
		o.Header("h1", TypeString, nil, nil).
			Tag("tag1").
			QueryObject(&q{Q3: 5}, nil).
			Path("p1", TypeInteger, web.Phrase("lang"), nil).
			Body(&object{}, true, web.Phrase("lang"), nil).
			Response("200", 5, web.Phrase("desc"), nil).
			Desc(nil, web.Phrase("lang"))
	})
	m.Middleware(nil, http.MethodGet, "/path/{p1}/abc", "")

	o := d.paths["/path/{p1}/abc"].Operations["GET"]
	a.NotNil(o).
		True(o.RequestBody.Ignorable).
		Equal(o.Description, web.Phrase("lang")).
		Length(o.Paths, 0).
		Length(o.Queries, 3).
		Equal(o.Queries[2].Schema.Default, 5).
		NotNil(o.RequestBody.Body.Type, TypeObject).
		Length(d.paths["/path/{p1}/abc"].Paths, 1)

	m = d.API(func(o *Operation) {
		o.Tag("tag1").
			Header("h1", TypeString, nil, nil).
			Path("p1", TypeInteger, web.Phrase("lang"), nil).
			Body(&object{}, false, nil, nil)
	})
	a.PanicString(func() {
		m.Middleware(nil, http.MethodGet, "/path/{p}/abc", "")
	}, "路径参数 p1 不存在于路径")
}

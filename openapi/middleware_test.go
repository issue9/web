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

func TestDocument_API(t *testing.T) {
	a := assert.New(t, false)
	ss := newServer(a)
	d := New(ss, web.Phrase("title"))

	m := d.API(func(o *Operation) {
		o.Header("h1", TypeString, nil, nil).
			Tag("tag1").
			QueryObject(&q{}, nil).
			Path("p1", TypeInteger, web.Phrase("lang"), nil).
			Body(&object{}, true, web.Phrase("lang"), nil).
			Response("200", 5, web.Phrase("desc"), nil).
			Description = web.Phrase("lang")
	})
	m.Middleware(nil, http.MethodGet, "/path/{p1}/abc", "")

	o := d.paths["/path/{p1}/abc"].Operations["GET"]
	a.NotNil(o).
		True(o.RequestBody.Ignorable).
		Equal(o.Description, web.Phrase("lang")).
		Length(o.Paths, 0).
		Length(o.Queries, 3).
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

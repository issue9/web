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

var _ web.Middleware = &APIMiddleware{}

type q struct {
	Q1 string `query:"q1"`
	Q2 int    `query:"q2"`
	Q3 int
}

func TestAPIMiddleware(t *testing.T) {
	a := assert.New(t, false)
	d := New("1.0.0", web.Phrase("desc"))

	m := d.API("tag1").
		Header("h1", TypeString, nil).
		QueryObject(&q{}).
		Path("p1", TypeInteger, web.Phrase("lang")).
		Body(&object{})
	a.NotPanic(func() {
		m.Middleware(nil, http.MethodGet, "/path/{p1}/abc", "")

		o := d.paths["/path/{p1}/abc"].Operations["GET"]
		a.NotNil(o).
			Length(o.Paths, 1).
			Length(o.Queries, 3).
			NotNil(o.RequestBody.Body.Type, TypeObject)
	})

	m = d.API("tag1").Header("h1", TypeString, nil).Path("p1", TypeInteger, web.Phrase("lang")).Body(&object{})
	a.PanicString(func() {
		m.Middleware(nil, http.MethodGet, "/path/{p}/abc", "")
	}, "路径参数 p1 不存在于路径")
}

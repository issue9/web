// SPDX-License-Identifier: MIT

package context

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

func newContextWithQuery(a *assert.Assertion, path string) *Context {
	r := httptest.NewRequest(http.MethodGet, path, bytes.NewBufferString("123"))
	r.Header.Set("Accept", "*/*")
	w := httptest.NewRecorder()
	ctx := newServer(a).newContext(w, r)

	return ctx
}

func TestQueries_Int(t *testing.T) {
	a := assert.New(t)
	ctx := newContextWithQuery(a, "/queries/int?i1=1&i2=2&str=str")
	q := ctx.Queries()

	a.Equal(q.Int("i1", 9), 1)
	a.Equal(q.Int("i2", 9), 2)
	a.Equal(q.Int("i3", 9), 9)

	// 无法转换，会返回默认值，且添加错误信息
	a.Equal(q.Int("str", 3), 3)
	a.True(q.HasErrors())
}

func TestQueries_Int64(t *testing.T) {
	a := assert.New(t)
	ctx := newContextWithQuery(a, "/queries/int64?i1=1&i2=2&str=str")
	q := ctx.Queries()

	a.Equal(q.Int64("i1", 9), 1)
	a.Equal(q.Int64("i2", 9), 2)
	a.Equal(q.Int64("i3", 9), 9)

	// 无法转换，会返回默认值，且添加错误信息
	a.Equal(q.Int64("str", 3), 3)
	a.Equal(len(q.errors), 1)
}

func TestQueries_String(t *testing.T) {
	a := assert.New(t)
	ctx := newContextWithQuery(a, "/queries/string?s1=1&s2=2")
	q := ctx.Queries()

	a.Equal(q.String("s1", "9"), "1")
	a.Equal(q.String("s2", "9"), "2")
	a.Equal(q.String("s3", "9"), "9")

}

func TestQueries_Bool(t *testing.T) {
	a := assert.New(t)
	ctx := newContextWithQuery(a, "/queries/bool?b1=true&b2=true&str=str")
	q := ctx.Queries()

	a.True(q.Bool("b1", false))
	a.True(q.Bool("b2", false))
	a.False(q.Bool("b3", false))

	// 无法转换，会返回默认值，且添加错误信息
	a.False(q.Bool("str", false))
	a.Equal(len(q.Errors()), 1)
}

func TestQueries_Float64(t *testing.T) {
	a := assert.New(t)
	ctx := newContextWithQuery(a, "/queries/float64?i1=1.1&i2=2&str=str")
	q := ctx.Queries()

	a.Equal(q.Float64("i1", 9.9), 1.1)
	a.Equal(q.Float64("i2", 9.9), 2)
	a.Equal(q.Float64("i3", 9.9), 9.9)

	// 无法转换，会返回默认值，且添加错误信息
	a.Equal(q.Float64("str", 3), 3)
	a.True(q.HasErrors())
}

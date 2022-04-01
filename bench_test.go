// SPDX-License-Identifier: MIT

package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"

	"github.com/issue9/web/serialization/text"
	"github.com/issue9/web/serialization/text/testobject"
	"github.com/issue9/web/server/servertest"
)

func BenchmarkObject(b *testing.B) {
	a := assert.New(b, false)
	s := servertest.NewServer(a, nil)
	o := &testobject.TextObject{}

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := rest.Post(a, "/path", nil).
			Header("Accept", text.Mimetype).
			Header("content-type", text.Mimetype).
			Request()
		ctx := s.NewContext(w, r)
		Object(http.StatusTeapot, o, nil).Apply(ctx)
	}
}

func BenchmarkObject_withHeader(b *testing.B) {
	a := assert.New(b, false)
	s := servertest.NewServer(a, nil)
	o := &testobject.TextObject{}

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := rest.Post(a, "/path", nil).
			Header("Accept", text.Mimetype).
			Header("content-type", text.Mimetype).
			Request()
		ctx := s.NewContext(w, r)
		Object(http.StatusTeapot, o, map[string]string{"Location": "https://example.com"}).Apply(ctx)
	}
}

// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package context

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/web/encoding"
	"github.com/issue9/web/encoding/test"
)

// 一次普通的 POST 请求过程
func BenchmarkPost(b *testing.B) {
	a := assert.New(b)
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/path", bytes.NewBufferString("request,15"))
		r.Header.Set("Accept", encoding.DefaultMimeType)
		ctx := New(w, r)

		obj := &test.TextObject{}
		a.True(ctx.Read(obj))

		obj.Age++
		obj.Name = "response"
		ctx.Render(http.StatusCreated, obj, nil)
		a.Equal(w.Body.String(), "response,16")
	}
}

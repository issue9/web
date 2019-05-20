// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package context

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

func TestResult(t *testing.T) {
	a := assert.New(t)

	r := httptest.NewRequest(http.MethodGet, "/path", bytes.NewBufferString("123"))
	r.Header.Set("Accept", "*/*")
	w := httptest.NewRecorder()
	ctx := newContext(a, w, r, nil, nil)
	ctx.App.AddMessages(400, map[int]string{
		40010: "40010",
		40011: "40011",
	})

	rslt := ctx.NewResultWithDetail(40010, map[string]string{
		"k1": "v1",
		"k2": "v2",
	})
	a.True(rslt.HasDetail())

	rslt.Render()
	a.NotEmpty(w.Body.String())
}

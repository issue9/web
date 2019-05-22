// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package context

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

func TestResult(t *testing.T) {
	a := assert.New(t)

	r := httptest.NewRequest(http.MethodGet, "/path", bytes.NewBufferString("123"))
	r.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()
	ctx := &Context{
		App: newApp(a),

		Response:       w,
		Request:        r,
		OutputCharset:  nil,
		OutputMimetype: json.Marshal,

		InputCharset:  nil,
		InputMimetype: json.Unmarshal,
	}
	ctx.App.AddMessages(http.StatusBadRequest, map[int]string{
		40010: "40010",
		40011: "40011",
	})

	rslt := ctx.NewResultWithDetail(40010, map[string]string{
		"k1": "v1",
	})
	a.True(rslt.HasDetail())

	rslt.Render()
	a.Equal(w.Body.String(), `{"message":"40010","code":40010,"detail":[{"field":"k1","message":"v1"}]}`)
}

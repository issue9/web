// SPDX-License-Identifier: MIT

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
		builder: newBuilder(a),

		Response:       w,
		Request:        r,
		OutputCharset:  nil,
		OutputMimetype: json.Marshal,

		InputCharset:  nil,
		InputMimetype: json.Unmarshal,
	}
	ctx.builder.Results().AddMessages(http.StatusBadRequest, map[int]string{
		40010: "40010",
		40011: "40011",
	})

	rslt := ctx.NewResultWithDetail(40010, map[string]string{
		"k1": "v1",
	})
	a.True(rslt.HasDetail())

	rslt.Render()
	a.Equal(w.Body.String(), `{"message":"40010","code":40010,"fields":[{"name":"k1","message":["v1"]}]}`)
}

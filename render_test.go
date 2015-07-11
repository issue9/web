// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

func TestRenderJson(t *testing.T) {
	a := assert.New(t)
	test := func(v interface{}, data string) {
		w := httptest.NewRecorder()
		a.NotNil(w)
		RenderJson(w, 200, v, map[string]string{"k": "v"})
		a.Equal(w.Body.String(), data)
		a.Equal(w.Header().Get("k"), "v")
	}

	// 特殊的值转换
	test(nil, `{}`)
	test("", `{}`)
	test(0, `0`)

	// 字符串类直接输出内容
	test(`{"abc":"abc"}`, `{"abc":"abc"}`)
	test([]byte(`{"abc":"abc"}`), `{"abc":"abc"}`)
	test([]rune(`{"abc":"abc"}`), `{"abc":"abc"}`)

	// map转换
	test(map[string]string{"abc": "abc"}, `{"abc":"abc"}`)
	test(map[string]int{"num": 1}, `{"num":1}`)

	err := errors.New("abc")
	test(err, "{}")
}

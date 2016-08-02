// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
)

func TestResultFields(t *testing.T) {
	a := assert.New(t)
	allow := []string{"col1", "col2", "col3"}

	//
	//ResultFields(r *http.Request, allow []string) ([]string, bool)

	r, err := http.NewRequest(http.MethodPut, "/test", nil)
	a.NotError(err).NotNil(r)

	// 非 GET 方法
	ret, ok := ResultFields(r, allow)
	a.False(ok).Nil(ret)

	// 非 GET 方法，即使设置 Method 方法
	r.Header.Set("X-Result-Fields", "col1, col2")
	ret, ok = ResultFields(r, allow)
	a.False(ok).Nil(ret)

	// 指定的字段都是允许的字段
	r.Method = http.MethodGet
	ret, ok = ResultFields(r, allow)
	a.True(ok).Equal([]string{"col1", "col2"}, ret)

	// 包含不允许的字段
	r.Header.Set("X-Result-Fields", "col1,col2, col100 ,col101")
	ret, ok = ResultFields(r, allow)
	a.False(ok).Equal([]string{"col100", "col101"}, ret)

	// 未指定 X-Result-Fields
	r.Header.Del("X-Result-Fields")
	ret, ok = ResultFields(r, allow)
	a.True(ok).Nil(nil)
}

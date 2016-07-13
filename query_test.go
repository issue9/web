// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
)

func TestQueryString(t *testing.T) {
	a := assert.New(t)

	r, err := http.NewRequest("GET", "/index.php?arg1=1&arg2=2", nil)
	a.NotError(err).NotNil(r)
	a.Equal("1", QueryString(r, "arg1", "2"))
	a.Equal("2", QueryString(r, "arg2", "1"))
	a.Equal("3", QueryString(r, "arg3", "3")) // 不存在的查询参数
}

func TestQueryInt64(t *testing.T) {
	a := assert.New(t)

	r, err := http.NewRequest("GET", "/index.php?arg1=1&arg2=two", nil)
	a.NotError(err).NotNil(r)

	ret, ok := QueryInt64(r, "arg1", 2)
	a.True(ok).Equal(ret, 1)

	// 无法转换
	ret, ok = QueryInt64(r, "arg2", 2)
	a.False(ok).Equal(ret, 2)

	// 不存在的查询参数
	ret, ok = QueryInt64(r, "arg3", 3)
	a.True(ok).Equal(ret, 3)
}

func TestQueryInt(t *testing.T) {
	a := assert.New(t)

	r, err := http.NewRequest("GET", "/index.php?arg1=1&arg2=two", nil)
	a.NotError(err).NotNil(r)

	ret, ok := QueryInt(r, "arg1", 2)
	a.True(ok).Equal(ret, 1)

	// 无法转换
	ret, ok = QueryInt(r, "arg2", 2)
	a.False(ok).Equal(ret, 2)

	// 不存在的查询参数
	ret, ok = QueryInt(r, "arg3", 3)
	a.True(ok).Equal(ret, 3)
}

func TestQueryFloat64(t *testing.T) {
	a := assert.New(t)

	r, err := http.NewRequest("GET", "/index.php?arg1=0.1&arg2=two", nil)
	a.NotError(err).NotNil(r)

	ret, ok := QueryFloat64(r, "arg1", 2)
	a.True(ok).Equal(ret, 0.1)

	// 无法转换
	ret, ok = QueryFloat64(r, "arg2", 0.2)
	a.False(ok).Equal(ret, 0.2)

	// 不存在的查询参数
	ret, ok = QueryFloat64(r, "arg3", 0.3)
	a.True(ok).Equal(ret, 0.3)
}

func TestQueryBool(t *testing.T) {
	a := assert.New(t)

	r, err := http.NewRequest("GET", "/index.php?arg1=true&arg2=off", nil)
	a.NotError(err).NotNil(r)

	ret, ok := QueryBool(r, "arg1", true)
	a.True(ok).True(ret)

	// 无法转换
	ret, ok = QueryBool(r, "arg2", true)
	a.False(ok).True(ret)

	// 不存在的查询参数
	ret, ok = QueryBool(r, "arg3", false)
	a.True(ok).False(ret)
}

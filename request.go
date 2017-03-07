// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"net/http"

	"github.com/issue9/web/contentype"
	"github.com/issue9/web/request"
)

// NewParam 声明一个新的 request.Param 实例
func NewParam(r *http.Request) *request.Param {
	return request.NewParam(r)
}

// NewQuery 声明一个新的 request.Query 实例
func NewQuery(r *http.Request) *request.Query {
	return request.NewQuery(r)
}

// ParamID 获取地址参数中表示 ID 的值。相对于 int64，但该值必须大于 0。
// 当出错时，第二个参数返回 false。
func ParamID(w http.ResponseWriter, r *http.Request, key string, code int, render contentype.Renderer) (int64, bool) {
	p := NewParam(r)
	id := p.Int64(key)
	rslt := p.Result(code)

	if rslt.HasDetail() {
		rslt.Render(render, w)
		return id, false
	}

	if id <= 0 {
		rslt.Add("id", "必须大于零")
		rslt.Render(render, w)
		return id, false
	}

	return id, true
}

// ParamInt64 取地址参数中的 int64 值
func ParamInt64(w http.ResponseWriter, r *http.Request, key string, code int, render contentype.Renderer) (int64, bool) {
	p := NewParam(r)
	id := p.Int64(key)

	if p.OK(code, render, w) {
		return id, false
	}

	return id, true
}

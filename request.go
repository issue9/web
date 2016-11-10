// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"net/http"

	"github.com/issue9/web/request"
)

// Param 获取一个 request.Param 实例，用于查询路径中的参数
func Param(r *http.Request, abortOnError bool) *request.Param {
	p, err := request.NewParam(r, abortOnError)
	if err != nil {
		Debug(r, err)
	}

	return p
}

// Query 获取一个 request.Query 实例，用于查询路径中的查询参数
func Query(r *http.Request, abortOnError bool) *request.Query {
	return request.NewQuery(r, abortOnError)
}

// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"net/http"

	"github.com/issue9/middleware/compress"

	"github.com/issue9/web/context"
	"github.com/issue9/web/internal/app"
	"github.com/issue9/web/internal/errors"
	"github.com/issue9/web/module"
	"github.com/issue9/web/result"
)

type (
	// Context 等同于 context.Context，方便调用者使用
	Context = context.Context

	// Result 等同于 result.Result，方便调用者使用
	Result = result.Result

	// Module 等同于 module.Module，方便调用者使用
	Module = module.Module
)

// AddErrorHandler 添加对错误状态码的处理方式。
//
// status 表示状态码，如果为 0，则表示所有未指定的状态码。
func AddErrorHandler(f func(http.ResponseWriter, int), status ...int) error {
	return errors.AddErrorHandler(f, status...)
}

// SetErrorHandler 设置指定状态码对应的处理函数
//
// 有则修改，没有则添加
//
// status 表示状态码，如果为 0，则表示所有未指定的状态码。
func SetErrorHandler(f func(http.ResponseWriter, int), status ...int) {
	errors.SetErrorHandler(f, status...)
}

// AddCompress 添加压缩方法。框架本身已经指定了 gzip 和 deflate 两种方法。
//
// NOTE: 只有在 web.Init() 之前调用才能启作用。
func AddCompress(name string, f compress.WriterFunc) error {
	return app.AddCompress(name, f)
}

// SetCompress 修改或是添加压缩方法。
//
// NOTE: 只有在 web.Init() 之前调用才能启作用。
func SetCompress(name string, f compress.WriterFunc) {
	app.SetCompress(name, f)
}

// NewContext 根据当前配置，生成 context.Context 对象，若是出错则 panic
func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return defaultApp.NewContext(w, r)
}

// NewResult 生成一个 *result.Result 对象
func NewResult(code int) *Result {
	return &result.Result{Code: code}
}

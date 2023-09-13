// SPDX-License-Identifier: MIT

// Package dev Development 环境下的功能实现
package dev

import (
	"path/filepath"
	"strings"

	"github.com/issue9/mux/v7"

	"github.com/issue9/web"
)

func Filename(f string) string {
	ext := filepath.Ext(f)
	base := strings.TrimSuffix(f, ext)

	// 用 _development 而不是 .development，防止在文件没有扩展名的情况下改变了文件的扩展名。
	return base + "_development" + ext
}

// DebugRouter 在非生产环境下为 r 提供一组测试用的 API
//
// path 测试路径；
// id 在取地址参数出错时的 problem id；
func DebugRouter(r *web.Router, path, id string) {
	if path == "" {
		panic("path 不能为空")
	}
	if path[0] != '/' {
		path = "/" + path
	}
	if path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	r.Any(path+"/{path}", func(ctx *web.Context) web.Responser {
		p, resp := ctx.PathString("path", id)
		if resp != nil {
			return resp
		}

		if err := mux.Debug(p, ctx, ctx.Request()); err != nil {
			return ctx.Error(err, "")
		}
		return nil
	})
}

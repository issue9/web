// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"net/http"

	"github.com/issue9/web/contentype"
)

// 编码解码工具
var defaultContentType contentype.ContentTyper

// Render 用于将 v 转换成相应的编码数据并写入到 w 中
//
// 若输出带报头的内容，可调用 ContentType().Render()
func Render(w http.ResponseWriter, code int, v interface{}) {
	defaultContentType.Render(w, code, v, nil)
}

// Read 用于将 r 中的 body 当作一个指定格式的数据读取到 v 中。
//
// 返回值指定是否出错。若出错，会在函数体中指定出错信息，并将错误代码写入报头。
func Read(w http.ResponseWriter, r *http.Request, v interface{}) bool {
	return defaultContentType.Read(w, r, v)
}

// ContentType 返回接口的实例
func ContentType() contentype.ContentTyper {
	return defaultContentType
}

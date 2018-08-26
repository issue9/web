// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package compress

import (
	"fmt"
	"log"
	"net/http"

	xcompress "github.com/issue9/middleware/compress"
)

var funcs = map[string]xcompress.WriterFunc{
	"gizp":    xcompress.NewGzip,
	"deflate": xcompress.NewDeflate,
}

// Add 添加压缩方法。框架本身已经指定了 gzip 和 deflate 两种方法。
func Add(name string, f xcompress.WriterFunc) error {
	if _, found := funcs[name]; found {
		return fmt.Errorf("已经存在同名 %s 的压缩函数", name)
	}

	funcs[name] = f
	return nil
}

// Set 修改或是添加压缩方法。
func Set(name string, f xcompress.WriterFunc) {
	funcs[name] = f
}

// Handler 返回封装后的 http.handler 实例
func Handler(h http.Handler, errlog *log.Logger) http.Handler {
	return xcompress.New(h, errlog, funcs)
}

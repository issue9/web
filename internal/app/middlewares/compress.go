// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package middlewares

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

// AddCompress 添加压缩方法。框架本身已经指定了 gzip 和 deflate 两种方法。
func AddCompress(name string, f xcompress.WriterFunc) error {
	if _, found := funcs[name]; found {
		return fmt.Errorf("已经存在同名 %s 的压缩函数", name)
	}

	funcs[name] = f
	return nil
}

// SetCompress 修改或是添加压缩方法。
func SetCompress(name string, f xcompress.WriterFunc) {
	funcs[name] = f
}

func compress(h http.Handler, errlog *log.Logger) http.Handler {
	opt := &xcompress.Options{
		Funcs: funcs,
		Types: []string{},
		Size:  0,
	}
	return xcompress.New(h, opt)
}

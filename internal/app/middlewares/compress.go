// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package middlewares

import (
	"fmt"

	"github.com/issue9/middleware/compress"
)

var funcs = map[string]compress.WriterFunc{
	"gizp":    compress.NewGzip,
	"deflate": compress.NewDeflate,
}

// AddCompress 添加压缩方法。框架本身已经指定了 gzip 和 deflate 两种方法。
func AddCompress(name string, f compress.WriterFunc) error {
	if _, found := funcs[name]; found {
		return fmt.Errorf("已经存在同名 %s 的压缩函数", name)
	}

	funcs[name] = f
	return nil
}

// SetCompress 修改或是添加压缩方法。
func SetCompress(name string, f compress.WriterFunc) {
	funcs[name] = f
}

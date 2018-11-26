// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package context

import (
	"testing"

	"github.com/issue9/assert"
	"golang.org/x/text/encoding/simplifiedchinese"

	"github.com/issue9/web/encoding"
)

func TestAcceptCharset(t *testing.T) {
	a := assert.New(t)

	name, enc, err := acceptCharset(encoding.DefaultCharset)
	a.NotError(err).
		Equal(name, encoding.DefaultCharset).
		True(charsetIsNop(enc))

	name, enc, err = acceptCharset("")
	a.NotError(err).
		Equal(name, encoding.DefaultCharset).
		True(charsetIsNop(enc))

	// * 表示采用默认的编码
	name, enc, err = acceptCharset("*")
	a.NotError(err).
		Equal(name, encoding.DefaultCharset).
		True(charsetIsNop(enc))

	name, enc, err = acceptCharset("gbk")
	a.NotError(err).
		Equal(name, "gbk").
		Equal(enc, simplifiedchinese.GBK)

	// 传递一个非正规名称
	name, enc, err = acceptCharset("chinese")
	a.NotError(err).
		Equal(name, "gbk").
		Equal(enc, simplifiedchinese.GBK)

	// q 错解析错误
	name, enc, err = acceptCharset("utf-8;q=x.9,gbk;q=0.8")
	a.Error(err).
		Equal(name, "").
		Nil(enc)

	// 不支持的编码
	name, enc, err = acceptCharset("not-supported")
	a.Error(err).
		Empty(name).
		Nil(enc)
}

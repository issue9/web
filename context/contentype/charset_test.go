// SPDX-License-Identifier: MIT

package contentype

import (
	"testing"

	"github.com/issue9/assert"
	"golang.org/x/text/encoding/simplifiedchinese"
)

func TestAcceptCharset(t *testing.T) {
	a := assert.New(t)

	name, enc, err := AcceptCharset(DefaultCharset)
	a.NotError(err).
		Equal(name, DefaultCharset).
		True(CharsetIsNop(enc))

	name, enc, err = AcceptCharset("")
	a.NotError(err).
		Equal(name, DefaultCharset).
		True(CharsetIsNop(enc))

	// * 表示采用默认的编码
	name, enc, err = AcceptCharset("*")
	a.NotError(err).
		Equal(name, DefaultCharset).
		True(CharsetIsNop(enc))

	name, enc, err = AcceptCharset("gbk")
	a.NotError(err).
		Equal(name, "gbk").
		Equal(enc, simplifiedchinese.GBK)

	// 传递一个非正规名称
	name, enc, err = AcceptCharset("chinese")
	a.NotError(err).
		Equal(name, "gbk").
		Equal(enc, simplifiedchinese.GBK)

	// q 错解析错误
	name, enc, err = AcceptCharset("utf-8;q=x.9,gbk;q=0.8")
	a.NotError(err).
		Equal(name, "gbk").
		Equal(enc, simplifiedchinese.GBK)

	// 不支持的编码
	name, enc, err = AcceptCharset("not-supported")
	a.Error(err).
		Empty(name).
		Nil(enc)
}

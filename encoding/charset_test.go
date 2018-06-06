// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package encoding

import (
	"testing"

	"github.com/issue9/assert"
	xencoding "golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
)

func TestFindCharset(t *testing.T) {
	a := assert.New(t)

	a.Equal(len(charset), 1) // 有一条默认的字符集信息
	name, intf := findCharset(DefaultCharset)
	a.Equal(name, DefaultCharset).NotNil(intf)

	// 添加已存在的
	a.Equal(AddCharset(DefaultCharset, simplifiedchinese.GBK), ErrExists)
	a.Equal(len(charset), 1) // 添加没成功

	a.NotError(AddCharset("GBK", simplifiedchinese.GBK))
	a.Equal(len(charset), 2) // 添加成功
	name, intf = findCharset("gbk")
	a.Equal(name, "GBK").NotNil(intf)

	name, intf = findCharset("*")
	a.Equal(name, DefaultCharset).Equal(intf, xencoding.Nop)
}

func TestAcceptCharset(t *testing.T) {
	a := assert.New(t)

	name, enc, err := AcceptCharset(DefaultCharset)
	a.NotError(err).
		Equal(enc, xencoding.Nop).
		Equal(name, DefaultCharset)

	name, enc, err = AcceptCharset("")
	a.NotError(err).
		Equal(enc, xencoding.Nop).
		Equal(name, DefaultCharset)

	// * 不指定，需要用户自行决定其表示方式
	name, enc, err = AcceptCharset("*")
	a.NotError(err).Equal(name, DefaultCharset)

	name, enc, err = AcceptCharset("utf-8;q=x.9,gbk;q=0.8")
	a.Error(err)
	a.Equal(name, "")

	name, enc, err = AcceptCharset("not-supported")
	a.Error(err).
		Empty(name).
		Nil(enc)
}

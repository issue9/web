// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package encoding

import (
	"testing"

	"github.com/issue9/assert"
	"golang.org/x/text/encoding/simplifiedchinese"
)

func TestCharset(t *testing.T) {
	a := assert.New(t)

	a.Equal(len(charset), 1) // 有一条默认的字符集信息
	a.Nil(Charset("not exists"))
	a.NotNil(Charset(DefaultCharset))

	// 添加已存在的
	a.Equal(AddCharset(DefaultCharset, simplifiedchinese.GBK), ErrExists)
	a.Equal(len(charset), 1) // 添加没成功

	a.NotError(AddCharset("GBK", simplifiedchinese.GBK))
	a.Equal(len(charset), 2) // 添加没成功
	a.NotNil(Charset("GBK"))
}

// SPDX-License-Identifier: MIT

package contentype

import (
	"errors"

	"github.com/issue9/qheader"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/encoding/unicode"
)

// DefaultCharset 默认的字符集
const DefaultCharset = "utf-8"

var (
	errInvalidCharset          = errors.New("无效的字符集")
	errContentTypeMissMimetype = errors.New("content-type 不存在 mimetype 部分")
)

// CharsetIsNop 指定的编码是否不需要任何额外操作
func CharsetIsNop(enc encoding.Encoding) bool {
	return enc == nil || enc == unicode.UTF8 || enc == encoding.Nop
}

// AcceptCharset 根据 Accept-Charset 报头的内容获取其最值的字符集信息
//
// 传递 * 获取返回默认的字符集相关信息，即 utf-8
// 其它值则按值查找，或是在找不到时返回空值。
//
// 返回的 name 值可能会与 header 中指定的不一样，比如 gb_2312 会被转换成 gbk
func AcceptCharset(header string) (name string, enc encoding.Encoding, err error) {
	if header == "" || header == "*" {
		return DefaultCharset, nil, nil
	}

	accepts := qheader.Parse(header, "*")
	for _, apt := range accepts {
		enc, err = htmlindex.Get(apt.Value)
		if err != nil { // err != nil 表示未找到，继续查找
			continue
		}

		// 转换成官方的名称
		name, err = htmlindex.Name(enc)
		if err != nil {
			name = apt.Value // 不存在，直接使用用户上传的名称
		}

		return name, enc, nil
	}

	return "", nil, errInvalidCharset
}

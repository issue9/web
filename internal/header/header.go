// SPDX-License-Identifier: MIT

// Package header 与报头相关的处理方法
package header

import (
	"strings"
	"unicode"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
	xunicode "golang.org/x/text/encoding/unicode"
)

const UTF8Name = "utf-8"

// ParseWithParam 分析带参数的报头
//
// 比如 content-type 可能带字符集的参数：content-type: application/json; charset=utf-8。
// 只返回主值以及指定名称的参数，其它忽略。
//
// 与 mime.ParseMediaType 最大的不同在于不会返回除 param 指定外的其它参数，
// 所以理论上性能也会更好一些，且也不局限于 RFC1521 规定的 content-type 报头，
// 对于 Accept 等，也可以分段解析。
func ParseWithParam(header, param string) (mt, charset string) {
	t, ps, found := strings.Cut(header, ";")
	if !found {
		return header, ""
	}

	for len(ps) > 0 {
		var item string
		if index := strings.IndexByte(ps, ';'); index > -1 {
			item = ps[:index]
			ps = ps[index+1:]
		} else {
			item = ps
			ps = ps[:0]
		}

		var name, val string
		if index := strings.IndexByte(item, '='); index >= 0 {
			name = item[:index]
			val = item[index+1:]
		} else { // 只有名称没有值
			name = item
		}

		if strings.ToLower(strings.TrimSpace(name)) == param {
			return t, strings.ToLower(strings.TrimFunc(val, func(r rune) bool { return unicode.IsSpace(r) || r == '"' }))
		}
	}

	return t, ""
}

// AcceptCharset 根据 Accept-Charset 报头的内容获取其最值的字符集信息
//
// 传递 * 获取返回默认的字符集相关信息，即 utf-8
// 其它值则按值查找，或是在找不到时返回空值。
//
// 返回的 name 值可能会与 header 中指定的不一样，比如 gb_2312 会被转换成 gbk
func AcceptCharset(header string) (name string, enc encoding.Encoding) {
	if header == "" || header == "*" {
		return UTF8Name, nil
	}

	items := ParseQHeader(header, "*")
	for _, item := range items {
		if item.Err != nil {
			continue
		}

		var err error
		if enc, err = htmlindex.Get(item.Value); err != nil { // err != nil 表示未找到，继续查找
			continue
		}

		// 转换成官方的名称
		if name, err = htmlindex.Name(enc); err != nil { // 不存在，直接使用用户上传的名称
			name = item.Value
		}
		return name, enc
	}

	return "", nil
}

// CharsetIsNop 指定的编码是否不需要任何额外操作
func CharsetIsNop(enc encoding.Encoding) bool {
	return enc == nil || enc == xunicode.UTF8 || enc == encoding.Nop
}

func BuildContentType(mt, charset string) string {
	if mt == "" {
		panic("mt 不能为空")
	}
	if charset == "" {
		charset = UTF8Name
	}

	return mt + "; charset=" + charset
}
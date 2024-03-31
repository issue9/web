// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package qheader

import (
	"net/http"
	"strings"
	"unicode"

	"github.com/issue9/mux/v8/header"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
	xunicode "golang.org/x/text/encoding/unicode"
)

// 一些常用的报头值
const (
	Identity  = "identity"
	KeepAlive = "keep-alive"
)

// ParseWithParam 分析带参数的报头
//
// 比如 content-type 可能带字符集的参数：content-type: application/json; charset=utf-8。
// 只返回主值以及指定名称的参数，其它忽略。
//
// 与 mime.ParseMediaType 最大的不同在于不会返回除 param 指定外的其它参数，
// 所以理论上性能也会更好一些，且也不局限于 RFC1521 规定的 content-type 报头，
// 对于 Accept 等，也可以分段解析。
// param 可以为空，表示不需要解析任何参数。
// paramValue 的返回值一律被转换为小写。
func ParseWithParam(header, param string) (mt, paramValue string) {
	t, ps, found := strings.Cut(header, ";")
	if !found {
		return header, ""
	}

	for len(ps) > 0 && param != "" {
		item, p, _ := strings.Cut(ps, ";")
		ps = p

		name, val, _ := strings.Cut(item, "=")
		if strings.ToLower(strings.TrimSpace(name)) == param {
			return t, strings.ToLower(strings.TrimFunc(val, func(r rune) bool { return unicode.IsSpace(r) || r == '"' }))
		}
	}

	return t, ""
}

// ParseAcceptCharset 根据 Accept-Charset 报头的内容获取其最值的字符集信息
//
// 传递 * 获取返回默认的字符集相关信息，即 utf-8
// 其它值则按值查找，或是在找不到时返回空值。
//
// 返回的 name 值可能会与 header 中指定的不一样，比如 gb_2312 会被转换成 gbk
func ParseAcceptCharset(h string) (name string, enc encoding.Encoding) {
	if h == "" || h == "*" {
		return header.UTF8, nil
	}

	items := ParseQHeader(h, "*")
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
		charset = header.UTF8
	}

	return mt + "; charset=" + charset
}

func ClientIP(r *http.Request) string {
	ip := r.Header.Get(header.XForwardedFor)
	if index := strings.IndexByte(ip, ','); index > 0 {
		ip = ip[:index]
	}
	if ip == "" && r.RemoteAddr != "" {
		ip = r.RemoteAddr
	}
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}

	return strings.TrimSpace(ip)
}

// InitETag 初始化 ETag 报头
//
// etag 为服务端生成的新值，包含了双引号，但不包含弱验证的 W/ 前缀；
// 返回值表示是否可以反馈 304 给客户。
func InitETag(w http.ResponseWriter, r *http.Request, etag string, weak bool) bool {
	client := r.Header.Get(header.IfNoneMatch)

	eq := client == etag ||
		(weak && len(client) > 1 && client[2:] == etag) // W/ 开头

	if weak {
		etag = "W/" + etag
	}
	w.Header().Set(header.ETag, etag)

	return eq
}

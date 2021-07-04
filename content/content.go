// SPDX-License-Identifier: MIT

// Package content 与生成内容相关的功能
package content

import (
	"mime"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/message/catalog"
)

// Content 管理反馈给用户的数据
type Content struct {
	mimetypes      []*mimetype
	resultMessages map[int]*resultMessage
	resultBuilder  BuildResultFunc
	catalog        *catalog.Builder
}

// New 返回 *Content 实例
func New(builder BuildResultFunc) *Content {
	return &Content{
		mimetypes:      make([]*mimetype, 0, 10),
		resultMessages: make(map[int]*resultMessage, 20),
		resultBuilder:  builder,
		catalog:        catalog.NewBuilder(),
	}
}

// CharsetIsNop 指定的编码是否不需要任何额外操作
func CharsetIsNop(enc encoding.Encoding) bool {
	return enc == nil || enc == unicode.UTF8 || enc == encoding.Nop
}

// ParseContentType 从 content-type 中获取编码和字符集
//
// 若客户端传回的是空值，则会使用默认值代替。
//
// 返回值中，mimetype 一律返回小写的值，charset 则原样返回
//
// https://tools.ietf.org/html/rfc7231#section-3.1.1.1
func ParseContentType(v string) (mimetype, charset string, err error) {
	if v = strings.TrimSpace(v); v == "" {
		return DefaultMimetype, DefaultCharset, nil
	}

	mt, params, err := mime.ParseMediaType(v)
	if err != nil {
		return "", "", err
	}
	if charset = params["charset"]; charset == "" {
		charset = DefaultCharset
	}
	return mt, charset, nil
}

// BuildContentType 生成一个 content-type
//
// 若值为空，则会使用默认值代替
func BuildContentType(mt, charset string) string {
	if mt == "" {
		mt = DefaultMimetype
	}
	if charset == "" {
		charset = DefaultCharset
	}

	return mt + "; charset=" + charset
}

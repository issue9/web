// SPDX-License-Identifier: MIT

// Package content 与生成内容相关的功能
package content

import (
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web/serialization"
)

// Content 管理反馈给用户的数据
type Content struct {
	mimetypes      *serialization.Serialization
	resultMessages map[int]*resultMessage
	resultBuilder  BuildResultFunc
	catalog        *catalog.Builder
}

// New 返回 *Content 实例
func New(builder BuildResultFunc) *Content {
	return &Content{
		mimetypes:      serialization.New(10),
		resultMessages: make(map[int]*resultMessage, 20),
		resultBuilder:  builder,
		catalog:        catalog.NewBuilder(),
	}
}

// 指定的编码是否不需要任何额外操作
func charsetIsNop(enc encoding.Encoding) bool {
	return enc == nil || enc == unicode.UTF8 || enc == encoding.Nop
}

// 生成 content-type，若值为空，则会使用默认值代替。
func buildContentType(mt, charset string) string {
	if mt == "" {
		mt = DefaultMimetype
	}
	if charset == "" {
		charset = DefaultCharset
	}

	return mt + "; charset=" + charset
}

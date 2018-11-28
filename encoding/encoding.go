// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package encoding 提供了框架内对编码和字符集功能的支持。
package encoding

// DefaultCharset 默认的字符集
const DefaultCharset = "utf-8"

// BuildContentType 生成一个 content-type
//
// 若值为空，则会使用默认值代替
func BuildContentType(mimetype, charset string) string {
	if mimetype == "" {
		mimetype = DefaultMimeType
	}
	if charset == "" {
		charset = DefaultCharset
	}

	return mimetype + "; charset=" + charset
}

// SPDX-License-Identifier: MIT

package content

import (
	"fmt"
	"mime"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"

	"github.com/issue9/web/serialization"
)

// DefaultMimetype 默认的媒体类型
//
// 在不能获取输入和输出的媒体类型时，会采用此值作为其默认值。
const DefaultMimetype = "application/octet-stream"

// Mimetypes 管理 mimetype 的序列化操作
func (c *Content) Mimetypes() *serialization.Mimetypes { return c.mimetypes }

// conentType 从 content-type 报头解析出需要用到的解码函数
func (c *Content) conentType(header string) (serialization.UnmarshalFunc, encoding.Encoding, error) {
	var (
		mt      = DefaultMimetype
		charset = DefaultCharset
	)

	if header != "" {
		mts, params, err := mime.ParseMediaType(header)
		if err != nil {
			return nil, nil, err
		}
		mt = mts
		if charset = params["charset"]; charset == "" {
			charset = DefaultCharset
		}
	}

	f, found := c.Mimetypes().UnmarshalFunc(mt)
	if !found {
		return nil, nil, fmt.Errorf("未注册的解函数 %s", mt)
	}

	e, err := htmlindex.Get(charset)
	if err != nil {
		return nil, nil, err
	}

	return f, e, nil
}

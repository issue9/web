// SPDX-License-Identifier: MIT

package content

import (
	"encoding/json"
	"testing"

	"github.com/issue9/assert"
)

func TestContent_contentType(t *testing.T) {
	a := assert.New(t)

	mt := New(DefaultBuilder)
	a.NotNil(mt)

	f, e, err := mt.conentType(";;;")
	a.Error(err).Nil(f).Nil(e)

	// 不存在的 mimetype
	f, e, err = mt.conentType(buildContentType(DefaultMimetype, DefaultCharset))
	a.Error(err).Nil(f).Nil(e)

	mt.Mimetypes().Add(nil, json.Unmarshal, DefaultMimetype)
	f, e, err = mt.conentType(buildContentType(DefaultMimetype, DefaultCharset))
	a.NotError(err).NotNil(f).NotNil(e)

	// 无效的字符集名称
	f, e, err = mt.conentType(buildContentType(DefaultMimetype, "invalid-charset"))
	a.Error(err).Nil(f).Nil(e)
}

// SPDX-License-Identifier: MIT

package content

import (
	"encoding/json"
	"encoding/xml"
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/web/serialization"
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

func TestContent_unmarshal(t *testing.T) {
	a := assert.New(t)

	mt := New(DefaultBuilder)
	a.NotNil(mt)

	um, found := mt.unmarshal("")
	a.False(found).Nil(um)

	a.NotError(mt.Mimetypes().Add(json.Marshal, json.Unmarshal, DefaultMimetype))

	um, found = mt.unmarshal(DefaultMimetype)
	a.True(found).NotNil(um)

	// 未指定 mimetype
	um, found = mt.unmarshal("")
	a.False(found).Nil(um)

	// mimetype 无法找到
	um, found = mt.unmarshal("not-exists")
	a.False(found).Nil(um)

	// 空的 unmarshal
	a.NotError(mt.Mimetypes().Add(json.Marshal, nil, "empty"))
	um, found = mt.unmarshal("empty")
	a.True(found).Nil(um)
}

func TestContent_marshal(t *testing.T) {
	a := assert.New(t)
	mt := New(DefaultBuilder)

	name, marshal, found := mt.marshal(DefaultMimetype)
	a.False(found).
		Nil(marshal).
		Empty(name)

	name, marshal, found = mt.marshal("")
	a.False(found).
		Nil(marshal).
		Empty(name)

	a.NotError(mt.Mimetypes().Add(xml.Marshal, xml.Unmarshal, DefaultMimetype))
	a.NotError(mt.Mimetypes().Add(json.Marshal, json.Unmarshal, "text/plain"))
	a.NotError(mt.Mimetypes().Add(nil, nil, "empty"))

	name, marshal, found = mt.marshal(DefaultMimetype)
	a.True(found).
		Equal(marshal, serialization.MarshalFunc(xml.Marshal)).
		Equal(name, DefaultMimetype)

	a.NotError(mt.Mimetypes().Set(DefaultMimetype, json.Marshal, json.Unmarshal))
	name, marshal, found = mt.marshal(DefaultMimetype)
	a.True(found).
		Equal(marshal, serialization.MarshalFunc(json.Marshal)).
		Equal(name, DefaultMimetype)

	a.ErrorString(mt.Mimetypes().Set("not-exists", nil, nil), "未找到指定名称")

	// */* 如果指定了 DefaultMimetype，则必定是该值
	name, marshal, found = mt.marshal("*/*")
	a.True(found).
		Equal(marshal, serialization.MarshalFunc(json.Marshal)).
		Equal(name, DefaultMimetype)

	// 同 */*
	name, marshal, found = mt.marshal("")
	a.True(found).
		Equal(marshal, serialization.MarshalFunc(json.Marshal)).
		Equal(name, DefaultMimetype)

	name, marshal, found = mt.marshal("*/*,text/plain")
	a.True(found).
		Equal(marshal, serialization.MarshalFunc(json.Marshal)).
		Equal(name, "text/plain")

	name, marshal, found = mt.marshal("font/wottf;q=x.9")
	a.False(found).
		Empty(name).
		Nil(marshal)

	name, marshal, found = mt.marshal("font/wottf")
	a.False(found).
		Empty(name).
		Nil(marshal)

	// 匹配 empty
	name, marshal, found = mt.marshal("empty")
	a.True(found).
		Equal(name, "empty").
		Nil(marshal)
}

func TestContent_findMarshal(t *testing.T) {
	a := assert.New(t)
	mt := New(DefaultBuilder)

	a.NotError(mt.Mimetypes().Add(nil, nil, "text", "text/plain", "text/text"))
	a.NotError(mt.Mimetypes().Add(nil, nil, "application/aa"))
	a.NotError(mt.Mimetypes().Add(nil, nil, "application/bb"))

	name, _ := mt.findMarshal("text")
	a.Equal(name, "text")

	name, _ = mt.findMarshal("text/*")
	a.Equal(name, "text")

	name, _ = mt.findMarshal("application/*")
	a.Equal(name, "application/aa")

	// 第一条数据
	name, _ = mt.findMarshal("*/*")
	a.Equal(name, "text")

	// 第一条数据
	name, _ = mt.findMarshal("")
	a.Equal(name, "text")

	// DefaultMimetype 不影响 findMarshal
	a.NotError(mt.Mimetypes().Add(nil, nil, DefaultMimetype))
	name, _ = mt.findMarshal("*/*")
	a.Equal(name, "text")

	// 不存在
	name, _ = mt.findMarshal("xx/*")
	a.Empty(name)
}

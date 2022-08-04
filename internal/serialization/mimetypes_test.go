// SPDX-License-Identifier: MIT

package serialization

import (
	"encoding/json"
	"encoding/xml"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/localeutil"

	"github.com/issue9/web/serializer"
)

const testMimetype = "application/octet-stream"

func TestMimetypes_unmarshalFunc(t *testing.T) {
	a := assert.New(t, false)

	mt := NewMimetypes(10)
	a.NotNil(mt)

	um, found := mt.unmarshalFunc("")
	a.False(found).Nil(um)

	a.NotError(mt.Add(json.Marshal, json.Unmarshal, testMimetype))

	um, found = mt.unmarshalFunc(testMimetype)
	a.True(found).NotNil(um)

	// 未指定 mimetype
	um, found = mt.unmarshalFunc("")
	a.False(found).Nil(um)

	// mimetype 无法找到
	um, found = mt.unmarshalFunc("not-exists")
	a.False(found).Nil(um)

	// 空的 UnmarshalFunc
	a.NotError(mt.Add(json.Marshal, nil, "empty"))
	um, found = mt.unmarshalFunc("empty")
	a.True(found).Nil(um)
}

func TestMimetypes_ContentType(t *testing.T) {
	a := assert.New(t, false)

	mt := NewMimetypes(10)
	a.NotError(mt.Add(json.Marshal, json.Unmarshal, "application/octet-stream"))
	a.NotNil(mt)

	f, e, err := mt.ContentType(";;;")
	a.Error(err).Nil(f).Nil(e)

	// 不存在的 mimetype
	f, e, err = mt.ContentType("not-exists; charset=utf-8")
	a.Equal(err, localeutil.Error("not found serialization function for %s", "not-exists")).Nil(f).Nil(e)

	// charset=utf-8
	f, e, err = mt.ContentType("application/octet-stream; charset=utf-8")
	a.NotError(err).NotNil(f).Nil(e)

	// charset=UTF-8
	f, e, err = mt.ContentType("application/octet-stream; charset=UTF-8;p1=k1;p2=k2")
	a.NotError(err).NotNil(f).Nil(e)

	// charset=
	f, e, err = mt.ContentType("application/octet-stream; charset=")
	a.NotError(err).NotNil(f).Nil(e)

	// 没有 charset
	f, e, err = mt.ContentType("application/octet-stream;")
	a.NotError(err).NotNil(f).Nil(e)

	// 没有 ;charset
	f, e, err = mt.ContentType("application/octet-stream")
	a.NotError(err).NotNil(f).Nil(e)

	// 未指定 charset 参数
	f, e, err = mt.ContentType("application/octet-stream; invalid-params")
	a.NotError(err).NotNil(f).Nil(e)
}

func TestMimetypes_MarshalFunc(t *testing.T) {
	a := assert.New(t, false)
	mt := NewMimetypes(10)

	name, marshal, found := mt.MarshalFunc(testMimetype)
	a.False(found).
		Nil(marshal).
		Empty(name)

	name, marshal, found = mt.MarshalFunc("")
	a.False(found).
		Nil(marshal).
		Empty(name)

	a.NotError(mt.Add(xml.Marshal, xml.Unmarshal, testMimetype))
	a.NotError(mt.Add(json.Marshal, json.Unmarshal, "text/plain"))
	a.NotError(mt.Add(nil, nil, "empty"))

	name, marshal, found = mt.MarshalFunc(testMimetype)
	a.True(found).
		Equal(marshal, serializer.MarshalFunc(xml.Marshal)).
		Equal(name, testMimetype)

	mt.Set(testMimetype, json.Marshal, json.Unmarshal)
	name, marshal, found = mt.MarshalFunc(testMimetype)
	a.True(found).
		Equal(marshal, serializer.MarshalFunc(json.Marshal)).
		Equal(name, testMimetype)

	// */* 如果指定了 DefaultMimetype，则必定是该值
	name, marshal, found = mt.MarshalFunc("*/*")
	a.True(found).
		Equal(marshal, serializer.MarshalFunc(json.Marshal)).
		Equal(name, testMimetype)

	// 同 */*
	name, marshal, found = mt.MarshalFunc("")
	a.True(found).
		Equal(marshal, serializer.MarshalFunc(json.Marshal)).
		Equal(name, testMimetype)

	name, marshal, found = mt.MarshalFunc("*/*,text/plain")
	a.True(found).
		Equal(marshal, serializer.MarshalFunc(json.Marshal)).
		Equal(name, "text/plain")

	name, marshal, found = mt.MarshalFunc("font/wottf;q=x.9")
	a.False(found).
		Empty(name).
		Nil(marshal)

	name, marshal, found = mt.MarshalFunc("font/wottf")
	a.False(found).
		Empty(name).
		Nil(marshal)

	// 匹配 empty
	name, marshal, found = mt.MarshalFunc("empty")
	a.True(found).
		Equal(name, "empty").
		Nil(marshal)
}

func TestMimetypes_findMarshal(t *testing.T) {
	a := assert.New(t, false)
	mt := NewMimetypes(10)

	a.NotError(mt.Add(nil, nil, "text", "text/plain", "text/text"))
	a.NotError(mt.Add(nil, nil, "application/aa"))
	a.NotError(mt.Add(nil, nil, "application/bb"))

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
	a.NotError(mt.Add(nil, nil, testMimetype))
	name, _ = mt.findMarshal("*/*")
	a.Equal(name, "text")

	// 不存在
	name, _ = mt.findMarshal("xx/*")
	a.Empty(name)
}

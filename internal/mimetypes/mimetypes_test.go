// SPDX-License-Identifier: MIT

package mimetypes

import (
	"encoding/json"
	"encoding/xml"
	"testing"

	"github.com/issue9/assert/v2"

	"github.com/issue9/web/serialization"
)

const testMimetype = "application/octet-stream"

func TestMimetypes_UnmarshalFunc(t *testing.T) {
	a := assert.New(t, false)

	mt := New(10)
	a.NotNil(mt)

	um, found := mt.UnmarshalFunc("")
	a.False(found).Nil(um)

	a.NotError(mt.Add(json.Marshal, json.Unmarshal, testMimetype))

	um, found = mt.UnmarshalFunc(testMimetype)
	a.True(found).NotNil(um)

	// 未指定 mimetype
	um, found = mt.UnmarshalFunc("")
	a.False(found).Nil(um)

	// mimetype 无法找到
	um, found = mt.UnmarshalFunc("not-exists")
	a.False(found).Nil(um)

	// 空的 UnmarshalFunc
	a.NotError(mt.Add(json.Marshal, nil, "empty"))
	um, found = mt.UnmarshalFunc("empty")
	a.True(found).Nil(um)
}

func TestMimetypes_MarshalFunc(t *testing.T) {
	a := assert.New(t, false)
	mt := New(10)

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
		Equal(marshal, serialization.MarshalFunc(xml.Marshal)).
		Equal(name, testMimetype)

	mt.Set(testMimetype, json.Marshal, json.Unmarshal)
	name, marshal, found = mt.MarshalFunc(testMimetype)
	a.True(found).
		Equal(marshal, serialization.MarshalFunc(json.Marshal)).
		Equal(name, testMimetype)

	// */* 如果指定了 DefaultMimetype，则必定是该值
	name, marshal, found = mt.MarshalFunc("*/*")
	a.True(found).
		Equal(marshal, serialization.MarshalFunc(json.Marshal)).
		Equal(name, testMimetype)

	// 同 */*
	name, marshal, found = mt.MarshalFunc("")
	a.True(found).
		Equal(marshal, serialization.MarshalFunc(json.Marshal)).
		Equal(name, testMimetype)

	name, marshal, found = mt.MarshalFunc("*/*,text/plain")
	a.True(found).
		Equal(marshal, serialization.MarshalFunc(json.Marshal)).
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
	mt := New(10)

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

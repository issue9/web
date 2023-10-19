// SPDX-License-Identifier: MIT

package codec

import (
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/localeutil"

	"github.com/issue9/web/codec/mimetype/json"
	"github.com/issue9/web/codec/mimetype/xml"
)

const testMimetype = "application/octet-stream"

func TestCodec_Add(t *testing.T) {
	a := assert.New(t, false)
	mt := New()

	a.False(mt.exists(testMimetype)).
		Empty(mt.AcceptHeader())

	mt.AddMimetype(testMimetype, json.BuildMarshal, json.Unmarshal, "")
	a.True(mt.exists(testMimetype)).
		Equal(mt.AcceptHeader(), testMimetype)

	mt.AddMimetype("application/json", json.BuildMarshal, json.Unmarshal, "application/problem+json")
	a.True(mt.exists(testMimetype)).
		Equal(mt.AcceptHeader(), testMimetype+",application/json")

	mt.AddMimetype("application/xml", nil, nil, "application/problem+json")
	a.True(mt.exists(testMimetype)).
		Equal(mt.AcceptHeader(), testMimetype+",application/json")

	a.Panic(func() {
		mt.AddMimetype(testMimetype, json.BuildMarshal, json.Unmarshal, "")
	}, "已经存在同名 application/octet-stream 的编码方法")
}

func TestCodec_ContentType(t *testing.T) {
	a := assert.New(t, false)

	mt := New()
	mt.AddMimetype(testMimetype, json.BuildMarshal, json.Unmarshal, "")
	a.NotNil(mt)

	f, e, err := mt.ContentType(";;;")
	a.Error(err).Nil(f).Nil(e)

	// 不存在的 mimetype
	f, e, err = mt.ContentType("not-exists; charset=utf-8")
	a.Equal(err, localeutil.Error("not found serialization function for %s", "not-exists")).Nil(f).Nil(e)

	// charset=utf-8
	f, e, err = mt.ContentType("application/octet-stream; charset=utf-8")
	a.NotError(err).NotNil(f).Nil(e)

	// charset=gb2312
	f, e, err = mt.ContentType("application/octet-stream; charset=gb2312")
	a.NotError(err).NotNil(f).NotNil(e)

	// charset=not-exists
	f, e, err = mt.ContentType("application/octet-stream; charset=not-exists")
	a.Error(err).Nil(f).Nil(e)

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

func TestCodec_Accept(t *testing.T) {
	a := assert.New(t, false)
	mt := New()

	item := mt.Accept(testMimetype)
	a.Nil(item)

	item = mt.Accept("")
	a.Nil(item)

	mt.AddMimetype(testMimetype, xml.BuildMarshal, xml.Unmarshal, "")
	mt.AddMimetype("text/plain", json.BuildMarshal, json.Unmarshal, "text/plain+problem")
	mt.AddMimetype("empty", nil, nil, "")

	item = mt.Accept(testMimetype)
	a.NotNil(item).
		NotNil(item.MarshalBuilder).
		Equal(item.Name(false), testMimetype).
		Equal(item.Name(true), testMimetype)

	// */* 如果指定了 DefaultMimetype，则必定是该值
	item = mt.Accept("*/*")
	a.NotNil(item).
		NotNil(item.MarshalBuilder).
		Equal(item.Name(false), testMimetype)

	// 同 */*
	item = mt.Accept("")
	a.NotNil(item).
		NotNil(item.MarshalBuilder).
		Equal(item.Name(false), testMimetype)

	item = mt.Accept("*/*,text/plain")
	a.NotNil(item).
		NotNil(item.MarshalBuilder).
		Equal(item.Name(false), "text/plain").
		Equal(item.Name(true), "text/plain+problem")

	item = mt.Accept("font/wottf;q=x.9")
	a.Nil(item)

	item = mt.Accept("font/wottf")
	a.Nil(item)

	// 匹配 empty
	item = mt.Accept("empty")
	a.NotNil(item).
		Equal(item.Name(false), "empty").
		Nil(item.MarshalBuilder())
}

func TestCodec_findMarshal(t *testing.T) {
	a := assert.New(t, false)
	mt := New()

	mt.AddMimetype("text", nil, nil, "")
	mt.AddMimetype("text/plain", nil, nil, "")
	mt.AddMimetype("text/text", nil, nil, "")
	mt.AddMimetype("application/aa", nil, nil, "")
	mt.AddMimetype("application/bb", nil, nil, "application/problem+bb")

	item := mt.findMarshal("text")
	a.NotNil(item).Equal(item.Name(false), "text")

	item = mt.findMarshal("text/*")
	a.NotNil(item).Equal(item.Name(false), "text")

	item = mt.findMarshal("application/*")
	a.NotNil(item).Equal(item.Name(false), "application/aa")

	// 第一条数据
	item = mt.findMarshal("*/*")
	a.NotNil(item).Equal(item.Name(false), "text")

	// 第一条数据
	item = mt.findMarshal("")
	a.NotNil(item).Equal(item.Name(false), "text")

	// DefaultMimetype 不影响 findMarshal
	mt.AddMimetype(testMimetype, nil, nil, "")
	item = mt.findMarshal("*/*")
	a.NotNil(item).Equal(item.Name(false), "text")

	// 通过 problem 查找
	item = mt.findMarshal("application/problem+bb")
	a.NotNil(item).Equal(item.Name(false), "application/bb")

	// 不存在
	item = mt.findMarshal("xx/*")
	a.Nil(item)
}

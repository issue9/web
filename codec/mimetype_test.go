// SPDX-License-Identifier: MIT

package codec

import (
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/localeutil"

	"github.com/issue9/web"
	"github.com/issue9/web/codec/mimetype/json"
	"github.com/issue9/web/codec/mimetype/xml"
)

const testMimetype = "application/octet-stream"

var _ web.Accepter = &mimetype{}

func TestCodec_ContentType(t *testing.T) {
	a := assert.New(t, false)

	mt := New([]*Mimetype{
		{Name: testMimetype, MarshalBuilder: json.BuildMarshal, Unmarshal: json.Unmarshal, Problem: ""},
	}, DefaultCompressions())
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
	mt := New(nil, nil)
	a.NotNil(mt)

	item := mt.Accept(testMimetype)
	a.Nil(item)

	item = mt.Accept("")
	a.Nil(item)

	mt = New([]*Mimetype{
		{Name: testMimetype, MarshalBuilder: xml.BuildMarshal, Unmarshal: xml.Unmarshal, Problem: ""},
		{Name: "text/plain", MarshalBuilder: json.BuildMarshal, Unmarshal: json.Unmarshal, Problem: "text/plain+problem"},
		{Name: "empty", MarshalBuilder: nil, Unmarshal: nil, Problem: ""},
	}, BestSpeedCompressions())
	a.NotNil(mt)

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
	mm := New([]*Mimetype{
		{Name: "text", MarshalBuilder: nil, Unmarshal: nil, Problem: ""},
		{Name: "text/plain", MarshalBuilder: nil, Unmarshal: nil, Problem: ""},
		{Name: "text/text", MarshalBuilder: nil, Unmarshal: nil, Problem: ""},
		{Name: "application/aa", MarshalBuilder: nil, Unmarshal: nil, Problem: ""},
		{Name: "application/bb", MarshalBuilder: nil, Unmarshal: nil, Problem: "application/problem+bb"},
		{Name: testMimetype, MarshalBuilder: nil, Unmarshal: nil, Problem: ""},
	}, BestCompressionCompressions())
	a.NotNil(mm)
	mt := mm.(*codec)

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
	item = mt.findMarshal("*/*")
	a.NotNil(item).Equal(item.Name(false), "text")

	// 通过 problem 查找
	item = mt.findMarshal("application/problem+bb")
	a.NotNil(item).Equal(item.Name(false), "application/bb")

	// 不存在
	item = mt.findMarshal("xx/*")
	a.Nil(item)
}

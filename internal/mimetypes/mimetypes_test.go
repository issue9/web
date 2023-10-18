// SPDX-License-Identifier: MIT

package mimetypes

import (
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/localeutil"
	"github.com/issue9/web/serializer/json"
	"github.com/issue9/web/serializer/xml"
)

const testMimetype = "application/octet-stream"

func TestMimetypes_Add(t *testing.T) {
	a := assert.New(t, false)
	mt := New(10)

	a.False(mt.exists(testMimetype)).
		Empty(mt.AcceptHeader())

	mt.Add(testMimetype, json.BuildMarshal, json.Unmarshal, "")
	a.True(mt.exists(testMimetype)).
		Equal(mt.AcceptHeader(), testMimetype)

	mt.Add("application/json", json.BuildMarshal, json.Unmarshal, "application/problem+json")
	a.True(mt.exists(testMimetype)).
		Equal(mt.AcceptHeader(), testMimetype+",application/json")

	mt.Add("application/xml", nil, nil, "application/problem+json")
	a.True(mt.exists(testMimetype)).
		Equal(mt.AcceptHeader(), testMimetype+",application/json")

	a.Panic(func() {
		mt.Add(testMimetype, json.BuildMarshal, json.Unmarshal, "")
	}, "已经存在同名 application/octet-stream 的编码方法")
}

func TestMimetypes_ContentType(t *testing.T) {
	a := assert.New(t, false)

	mt := New(10)
	mt.Add(testMimetype, json.BuildMarshal, json.Unmarshal, "")
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

func TestMimetypes_Accept(t *testing.T) {
	a := assert.New(t, false)
	mt := New(10)

	item := mt.Accept(testMimetype)
	a.Nil(item)

	item = mt.Accept("")
	a.Nil(item)

	mt.Add(testMimetype, xml.BuildMarshal, xml.Unmarshal, "")
	mt.Add("text/plain", json.BuildMarshal, json.Unmarshal, "text/plain+problem")
	mt.Add("empty", nil, nil, "")

	item = mt.Accept(testMimetype)
	a.NotNil(item).
		NotNil(item.MarshalBuilder).
		Equal(item.Name, testMimetype).
		Equal(item.Problem, testMimetype)

	// */* 如果指定了 DefaultMimetype，则必定是该值
	item = mt.Accept("*/*")
	a.NotNil(item).
		NotNil(item.MarshalBuilder).
		Equal(item.Name, testMimetype)

	// 同 */*
	item = mt.Accept("")
	a.NotNil(item).
		NotNil(item.MarshalBuilder).
		Equal(item.Name, testMimetype)

	item = mt.Accept("*/*,text/plain")
	a.NotNil(item).
		NotNil(item.MarshalBuilder).
		Equal(item.Name, "text/plain").
		Equal(item.Problem, "text/plain+problem")

	item = mt.Accept("font/wottf;q=x.9")
	a.Nil(item)

	item = mt.Accept("font/wottf")
	a.Nil(item)

	// 匹配 empty
	item = mt.Accept("empty")
	a.NotNil(item).
		Equal(item.Name, "empty").
		Nil(item.MarshalBuilder)
}

func TestMimetypes_findMarshal(t *testing.T) {
	a := assert.New(t, false)
	mt := New(10)

	mt.Add("text", nil, nil, "")
	mt.Add("text/plain", nil, nil, "")
	mt.Add("text/text", nil, nil, "")
	mt.Add("application/aa", nil, nil, "")
	mt.Add("application/bb", nil, nil, "application/problem+bb")

	item := mt.findMarshal("text")
	a.NotNil(item).Equal(item.Name, "text")

	item = mt.findMarshal("text/*")
	a.NotNil(item).Equal(item.Name, "text")

	item = mt.findMarshal("application/*")
	a.NotNil(item).Equal(item.Name, "application/aa")

	// 第一条数据
	item = mt.findMarshal("*/*")
	a.NotNil(item).Equal(item.Name, "text")

	// 第一条数据
	item = mt.findMarshal("")
	a.NotNil(item).Equal(item.Name, "text")

	// DefaultMimetype 不影响 findMarshal
	mt.Add(testMimetype, nil, nil, "")
	item = mt.findMarshal("*/*")
	a.NotNil(item).Equal(item.Name, "text")

	// 通过 problem 查找
	item = mt.findMarshal("application/problem+bb")
	a.NotNil(item).Equal(item.Name, "application/bb")

	// 不存在
	item = mt.findMarshal("xx/*")
	a.Nil(item)
}

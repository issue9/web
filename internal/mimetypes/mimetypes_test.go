// SPDX-License-Identifier: MIT

package mimetypes

import (
	"encoding/json"
	"encoding/xml"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/localeutil"
)

type (
	marshalFunc   = func(any) ([]byte, error)
	unmarshalFunc = func([]byte, any) error
)

const testMimetype = "application/octet-stream"

func TestMimetypes_contentType(t *testing.T) {
	a := assert.New(t, false)

	mt := New[marshalFunc, unmarshalFunc]()
	mt.Add(testMimetype, json.Marshal, json.Unmarshal, "")
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
	mt := New[marshalFunc, unmarshalFunc]()

	item := mt.MarshalFunc(testMimetype)
	a.Nil(item)

	item = mt.MarshalFunc("")
	a.Nil(item)

	mt.Add(testMimetype, xml.Marshal, xml.Unmarshal, "")
	mt.Add("text/plain", json.Marshal, json.Unmarshal, "text/plain+problem")
	mt.Add("empty", nil, nil, "")

	item = mt.MarshalFunc(testMimetype)
	a.NotNil(item).
		Equal(item.Marshal, marshalFunc(xml.Marshal)).
		Equal(item.Name, testMimetype).
		Equal(item.Problem, testMimetype)

	mt.Set(testMimetype, json.Marshal, json.Unmarshal, "")
	item = mt.MarshalFunc(testMimetype)
	a.NotNil(item).
		Equal(item.Marshal, marshalFunc(json.Marshal)).
		Equal(item.Name, testMimetype)

	// */* 如果指定了 DefaultMimetype，则必定是该值
	item = mt.MarshalFunc("*/*")
	a.NotNil(item).
		Equal(item.Marshal, marshalFunc(json.Marshal)).
		Equal(item.Name, testMimetype)

	// 同 */*
	item = mt.MarshalFunc("")
	a.NotNil(item).
		Equal(item.Marshal, marshalFunc(json.Marshal)).
		Equal(item.Name, testMimetype)

	item = mt.MarshalFunc("*/*,text/plain")
	a.NotNil(item).
		Equal(item.Marshal, marshalFunc(json.Marshal)).
		Equal(item.Name, "text/plain").
		Equal(item.Problem, "text/plain+problem")

	item = mt.MarshalFunc("font/wottf;q=x.9")
	a.Nil(item)

	item = mt.MarshalFunc("font/wottf")
	a.Nil(item)

	// 匹配 empty
	item = mt.MarshalFunc("empty")
	a.NotNil(item).
		Equal(item.Name, "empty").
		Nil(item.Marshal)
}

func TestMimetypes_findMarshal(t *testing.T) {
	a := assert.New(t, false)
	mt := New[marshalFunc, unmarshalFunc]()

	mt.Add("text", nil, nil, "")
	mt.Add("text/plain", nil, nil, "")
	mt.Add("text/text", nil, nil, "")
	mt.Add("application/aa", nil, nil, "")
	mt.Add("application/bb", nil, nil, "")

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

	// 不存在
	item = mt.findMarshal("xx/*")
	a.Nil(item)
}

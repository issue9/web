// SPDX-License-Identifier: MIT

package server

import (
	"encoding/json"
	"encoding/xml"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/localeutil"
)

const testMimetype = "application/octet-stream"

func marshalTest(_ *Context, v any) ([]byte, error) {
	switch vv := v.(type) {
	case error:
		return nil, vv
	default:
		return nil, ErrUnsupported
	}
}

func unmarshalTest(bs []byte, v any) error {
	return ErrUnsupported
}

func TestMimetypes_contentType(t *testing.T) {
	a := assert.New(t, false)

	mt := newMimetypes()
	mt.Add("application/octet-stream", MarshalJSON, json.Unmarshal, "")
	a.NotNil(mt)

	f, e, err := mt.contentType(";;;")
	a.Error(err).Nil(f).Nil(e)

	// 不存在的 mimetype
	f, e, err = mt.contentType("not-exists; charset=utf-8")
	a.Equal(err, localeutil.Error("not found serialization function for %s", "not-exists")).Nil(f).Nil(e)

	// charset=utf-8
	f, e, err = mt.contentType("application/octet-stream; charset=utf-8")
	a.NotError(err).NotNil(f).Nil(e)

	// charset=UTF-8
	f, e, err = mt.contentType("application/octet-stream; charset=UTF-8;p1=k1;p2=k2")
	a.NotError(err).NotNil(f).Nil(e)

	// charset=
	f, e, err = mt.contentType("application/octet-stream; charset=")
	a.NotError(err).NotNil(f).Nil(e)

	// 没有 charset
	f, e, err = mt.contentType("application/octet-stream;")
	a.NotError(err).NotNil(f).Nil(e)

	// 没有 ;charset
	f, e, err = mt.contentType("application/octet-stream")
	a.NotError(err).NotNil(f).Nil(e)

	// 未指定 charset 参数
	f, e, err = mt.contentType("application/octet-stream; invalid-params")
	a.NotError(err).NotNil(f).Nil(e)
}

func TestMimetypes_marshalFunc(t *testing.T) {
	a := assert.New(t, false)
	mt := newMimetypes()

	item := mt.marshalFunc(testMimetype)
	a.Nil(item)

	item = mt.marshalFunc("")
	a.Nil(item)

	mt.Add(testMimetype, MarshalXML, xml.Unmarshal, "")
	mt.Add("text/plain", MarshalJSON, json.Unmarshal, "text/plain+problem")
	mt.Add("empty", nil, nil, "")

	item = mt.marshalFunc(testMimetype)
	a.NotNil(item).
		Equal(item.marshal, MarshalFunc(MarshalXML)).
		Equal(item.name, testMimetype).
		Equal(item.problem, testMimetype)

	mt.Set(testMimetype, MarshalJSON, json.Unmarshal, "")
	item = mt.marshalFunc(testMimetype)
	a.NotNil(item).
		Equal(item.marshal, MarshalFunc(MarshalJSON)).
		Equal(item.name, testMimetype)

	// */* 如果指定了 DefaultMimetype，则必定是该值
	item = mt.marshalFunc("*/*")
	a.NotNil(item).
		Equal(item.marshal, MarshalFunc(MarshalJSON)).
		Equal(item.name, testMimetype)

	// 同 */*
	item = mt.marshalFunc("")
	a.NotNil(item).
		Equal(item.marshal, MarshalFunc(MarshalJSON)).
		Equal(item.name, testMimetype)

	item = mt.marshalFunc("*/*,text/plain")
	a.NotNil(item).
		Equal(item.marshal, MarshalFunc(MarshalJSON)).
		Equal(item.name, "text/plain").
		Equal(item.problem, "text/plain+problem")

	item = mt.marshalFunc("font/wottf;q=x.9")
	a.Nil(item)

	item = mt.marshalFunc("font/wottf")
	a.Nil(item)

	// 匹配 empty
	item = mt.marshalFunc("empty")
	a.NotNil(item).
		Equal(item.name, "empty").
		Nil(item.marshal)
}

func TestMimetypes_findMarshal(t *testing.T) {
	a := assert.New(t, false)
	mt := newMimetypes()

	mt.Add("text", nil, nil, "")
	mt.Add("text/plain", nil, nil, "")
	mt.Add("text/text", nil, nil, "")
	mt.Add("application/aa", nil, nil, "")
	mt.Add("application/bb", nil, nil, "")

	item := mt.findMarshal("text")
	a.NotNil(item).Equal(item.name, "text")

	item = mt.findMarshal("text/*")
	a.NotNil(item).Equal(item.name, "text")

	item = mt.findMarshal("application/*")
	a.NotNil(item).Equal(item.name, "application/aa")

	// 第一条数据
	item = mt.findMarshal("*/*")
	a.NotNil(item).Equal(item.name, "text")

	// 第一条数据
	item = mt.findMarshal("")
	a.NotNil(item).Equal(item.name, "text")

	// DefaultMimetype 不影响 findMarshal
	mt.Add(testMimetype, nil, nil, "")
	item = mt.findMarshal("*/*")
	a.NotNil(item).Equal(item.name, "text")

	// 不存在
	item = mt.findMarshal("xx/*")
	a.Nil(item)
}

// SPDX-FileCopyrightText: 2018-2025 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/lzw"
	"encoding/json"
	"encoding/xml"
	"io"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/mux/v9/header"

	"github.com/issue9/web/compressor"
	"github.com/issue9/web/internal/qheader"
	"github.com/issue9/web/mimetype"
)

func marshalTest(_ *Context, v any) ([]byte, error) {
	switch vv := v.(type) {
	case error:
		return nil, vv
	default:
		return nil, mimetype.ErrUnsupported()
	}
}

func unmarshalTest(io.Reader, any) error { return mimetype.ErrUnsupported() }

func marshalJSON(_ *Context, v any) ([]byte, error) { return json.Marshal(v) }

func unmarshalJSON(r io.Reader, v any) error { return json.NewDecoder(r).Decode(v) }

func marshalXML(_ *Context, v any) ([]byte, error) { return xml.Marshal(v) }

func unmarshalXML(r io.Reader, v any) error { return xml.NewDecoder(r).Decode(v) }

func newCodec(a *assert.Assertion) *Codec {
	c := NewCodec()
	a.NotNil(c)

	c.AddCompressor(compressor.NewGzip(gzip.BestSpeed)).
		AddCompressor(compressor.NewDeflate(flate.DefaultCompression, nil)).
		//AddCompressor(nil).
		AddMimetype(header.JSON, marshalJSON, unmarshalJSON, "application/problem+json").
		AddMimetype(header.XML, marshalXML, unmarshalXML, "application/problem+xml").
		AddMimetype("application/test", marshalTest, unmarshalTest, "application/problem+test")

	return c
}

func TestCodec_AddMimetype(t *testing.T) {
	a := assert.New(t, false)
	c := NewCodec()

	a.PanicString(func() {
		c.AddMimetype("", nil, nil, "")
	}, "参数 name 不能为空")

	a.PanicString(func() {
		c.AddMimetype(header.JSON, nil, nil, "")
	}, "参数 m 不能为空")

	a.PanicString(func() {
		c.AddMimetype(header.JSON, marshalJSON, nil, "")
	}, "参数 u 不能为空")

	a.NotPanic(func() {
		c.AddMimetype(header.JSON, marshalJSON, unmarshalJSON, "")
	})

	a.PanicString(func() {
		c.AddMimetype(header.JSON, marshalJSON, unmarshalJSON, "")
	}, "存在重复的项 "+header.JSON)
}

func TestBuildCompression(t *testing.T) {
	a := assert.New(t, false)

	c := buildCompression(compressor.NewGzip(gzip.DefaultCompression), nil)
	a.True(c.wildcard).
		Length(c.types, 0).
		Length(c.wildcardSuffix, 0)

	c = buildCompression(compressor.NewGzip(gzip.DefaultCompression), []string{"text"})
	a.Equal(c.types, []string{"text"})

	c = buildCompression(compressor.NewGzip(gzip.DefaultCompression), []string{"text", "*"})
	a.Nil(c.types).
		True(c.wildcard).
		Nil(c.wildcardSuffix)
}

func TestCodec_contentEncoding(t *testing.T) {
	a := assert.New(t, false)

	e := NewCodec()
	a.NotNil(e)
	e.AddCompressor(compressor.NewLZW(lzw.LSB, 8), header.Plain, "application/*").
		AddCompressor(compressor.NewGzip(gzip.BestSpeed), header.Plain).
		AddCompressor(compressor.NewZstd(), "application/*")

	r := &bytes.Buffer{}
	rc, err := e.contentEncoding("zstd", r)
	a.NotError(err).NotNil(rc)

	r = bytes.NewBufferString("123")
	rc, err = e.contentEncoding("", r)
	a.NotError(err).NotNil(rc)
	data, err := io.ReadAll(rc)
	a.NotError(err).Equal(string(data), "123")
}

func TestCodec_acceptEncoding(t *testing.T) {
	a := assert.New(t, false)

	e := NewCodec()
	a.NotNil(e)
	e.AddCompressor(compressor.NewLZW(lzw.LSB, 8), header.Plain, "application/*").
		AddCompressor(compressor.NewGzip(gzip.DefaultCompression), header.Plain).
		AddCompressor(compressor.NewGzip(gzip.DefaultCompression), "application/*")

	a.Equal(e.acceptEncodingHeader, "compress,gzip")

	t.Run("一般", func(t *testing.T) {
		a := assert.New(t, false)
		b, notAccept := e.acceptEncoding(header.JSON, "gzip;q=0.9,br")
		a.False(notAccept).NotNil(b).Equal(b.Name(), "gzip")

		b, notAccept = e.acceptEncoding(header.JSON, "br,gzip")
		a.False(notAccept).NotNil(b).Equal(b.Name(), "gzip")

		b, notAccept = e.acceptEncoding(header.Plain, "gzip,br")
		a.False(notAccept).NotNil(b).Equal(b.Name(), "gzip")

		b, notAccept = e.acceptEncoding(header.Plain, "br")
		a.False(notAccept).Nil(b)

		b, notAccept = e.acceptEncoding(header.Plain, "")
		a.False(notAccept).Nil(b)
	})

	t.Run("header=*", func(t *testing.T) {
		a := assert.New(t, false)
		b, notAccept := e.acceptEncoding(header.XML, "*;q=0")
		a.True(notAccept).Nil(b)

		b, notAccept = e.acceptEncoding(header.XML, "*,br")
		a.False(notAccept).NotNil(b).Equal(b.Name(), "compress")

		b, notAccept = e.acceptEncoding(header.XML, "*,gzip")
		a.False(notAccept).NotNil(b).Equal(b.Name(), "compress")

		b, notAccept = e.acceptEncoding(header.XML, "*,gzip,compress") // gzip,compress 都排除了
		a.False(notAccept).Nil(b)
	})

	t.Run("header=identity", func(t *testing.T) {
		a := assert.New(t, false)
		b, notAccept := e.acceptEncoding(header.XML, "identity,gzip,br")
		a.False(notAccept).NotNil(b).Equal(b.Name(), "gzip")

		// 正常匹配
		b, notAccept = e.acceptEncoding(header.XML, "identity;q=0,gzip,br")
		a.False(notAccept).NotNil(b).Equal(b.Name(), "gzip")

		// 没有可匹配，选取第一个
		b, notAccept = e.acceptEncoding(header.XML, "identity;q=0,abc,def")
		a.False(notAccept).Nil(b)
	})
}

func TestCodec_contentType(t *testing.T) {
	a := assert.New(t, false)

	mt := NewCodec()
	a.NotNil(mt)
	mt.AddMimetype(header.OctetStream, marshalJSON, unmarshalJSON, "")

	f, e, err := mt.contentType(";;;")
	a.Error(err).Nil(f).Nil(e)

	// 不存在的 mimetype
	f, e, err = mt.contentType("not-exists; charset=utf-8")
	a.Equal(err, NewLocaleError("not found serialization function for %s", "not-exists")).Nil(f).Nil(e)

	// charset=utf-8
	f, e, err = mt.contentType(qheader.BuildContentType(header.OctetStream, header.UTF8))
	a.NotError(err).NotNil(f).Nil(e)

	// charset=gb2312
	f, e, err = mt.contentType(qheader.BuildContentType(header.OctetStream, "gb2312"))
	a.NotError(err).NotNil(f).NotNil(e)

	// charset=not-exists
	f, e, err = mt.contentType(qheader.BuildContentType(header.OctetStream, "not-exists"))
	a.Error(err).Nil(f).Nil(e)

	// charset=UTF-8
	f, e, err = mt.contentType("application/octet-stream; charset=UTF-8;p1=k1;p2=k2")
	a.NotError(err).NotNil(f).Nil(e)

	// charset=
	f, e, err = mt.contentType(qheader.BuildContentType(header.OctetStream, ""))
	a.NotError(err).NotNil(f).Nil(e)

	// 没有 charset
	f, e, err = mt.contentType("application/octet-stream;")
	a.NotError(err).NotNil(f).Nil(e)

	// 没有 ;charset
	f, e, err = mt.contentType(header.OctetStream)
	a.NotError(err).NotNil(f).Nil(e)

	// 未指定 charset 参数
	f, e, err = mt.contentType("application/octet-stream; invalid-params")
	a.NotError(err).NotNil(f).Nil(e)
}

func TestCodec_accept(t *testing.T) {
	a := assert.New(t, false)
	mt := NewCodec()
	a.NotNil(mt)

	item := mt.accept(header.JSON)
	a.Nil(item)

	item = mt.accept("")
	a.Nil(item)

	mt = NewCodec()
	a.NotNil(mt)
	mt.AddMimetype(header.JSON, marshalJSON, unmarshalJSON, "").
		AddMimetype(header.Plain, marshalXML, unmarshalXML, "text/plain+problem")

	item = mt.accept(header.JSON)
	a.NotNil(item).
		NotNil(item.Marshal).
		Equal(item.name(false), header.JSON).
		Equal(item.name(true), header.JSON)

	// */*
	item = mt.accept("*/*")
	a.NotNil(item).
		NotNil(item.Marshal).
		Equal(item.name(false), header.JSON)

	// 空参数，结果同 */*
	item = mt.accept("")
	a.NotNil(item).
		NotNil(item.Marshal).
		Equal(item.name(false), header.JSON)

	item = mt.accept("*/*,text/plain")
	a.NotNil(item).
		NotNil(item.Marshal).
		Equal(item.name(false), header.Plain).
		Equal(item.name(true), "text/plain+problem")

	item = mt.accept("font/wottf;q=x.9")
	a.Nil(item)

	item = mt.accept("font/wottf")
	a.Nil(item)
}

func TestCodec_findMarshal(t *testing.T) {
	a := assert.New(t, false)
	mm := NewCodec()
	a.NotNil(mm)
	mm.AddMimetype("text", marshalTest, unmarshalTest, "").
		AddMimetype("text/plain", marshalTest, unmarshalTest, "").
		AddMimetype("text/text", marshalTest, unmarshalTest, "").
		AddMimetype("application/aa", marshalTest, unmarshalTest, "").
		AddMimetype("application/bb", marshalTest, unmarshalTest, "application/problem+bb").
		AddMimetype(header.JSON, marshalTest, unmarshalTest, "")

	item := mm.findMarshal("text")
	a.NotNil(item).Equal(item.name(false), "text")

	item = mm.findMarshal("text/*")
	a.NotNil(item).Equal(item.name(false), "text")

	item = mm.findMarshal("application/*")
	a.NotNil(item).Equal(item.name(false), "application/aa")

	// 第一条数据
	item = mm.findMarshal("*/*")
	a.NotNil(item).Equal(item.name(false), "text")

	// 第一条数据
	item = mm.findMarshal("")
	a.NotNil(item).Equal(item.name(false), "text")

	// DefaultMimetype 不影响 findMarshal
	item = mm.findMarshal("*/*")
	a.NotNil(item).Equal(item.name(false), "text")

	// 通过 problem 查找
	item = mm.findMarshal("application/problem+bb")
	a.NotNil(item).Equal(item.name(false), "application/bb")

	// 不存在
	item = mm.findMarshal("xx/*")
	a.Nil(item)
}

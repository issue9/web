// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/lzw"
	"encoding/json/v2"
	"encoding/xml"
	"errors"
	"io"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/issue9/assert/v6"
	"github.com/issue9/localeutil"
	"github.com/issue9/mux/v10/header"
	"github.com/klauspost/compress/zstd"

	"github.com/issue9/web/internal/qheader"
)

type (
	gzipImpl    struct{}
	deflateImpl struct{}
	lzwImpl     struct{}
	zstdImpl    struct{}
	brotliImpl  struct{}

	mediaTypeImpl struct {
		normalName  string
		problemName string
		marshal     MarshalFunc
		unmarshal   UnmarshalFunc
	}
)

func newGzip() Compressor                                       { return &gzipImpl{} }
func (gzipImpl) Name() string                                   { return "gzip" }
func (gzipImpl) NewDecoder(r io.Reader) (io.ReadCloser, error)  { return gzip.NewReader(r) }
func (gzipImpl) NewEncoder(w io.Writer) (io.WriteCloser, error) { return gzip.NewWriter(w), nil }

func newLZW() Compressor     { return &lzwImpl{} }
func (lzwImpl) Name() string { return "compress" }
func (c lzwImpl) NewDecoder(r io.Reader) (io.ReadCloser, error) {
	return lzw.NewReader(r, lzw.LSB, 8), nil
}
func (c lzwImpl) NewEncoder(w io.Writer) (io.WriteCloser, error) {
	return lzw.NewWriter(w, lzw.LSB, 8), nil
}

func newZstd() Compressor     { return &zstdImpl{} }
func (zstdImpl) Name() string { return "zstd" }
func (zstdImpl) NewDecoder(r io.Reader) (io.ReadCloser, error) {
	rr, err := zstd.NewReader(r)
	if err != nil {
		return nil, err
	}
	return io.NopCloser(rr), nil
}
func (zstdImpl) NewEncoder(w io.Writer) (io.WriteCloser, error) { return zstd.NewWriter(w) }

func newBrotli() Compressor     { return &brotliImpl{} }
func (brotliImpl) Name() string { return "br" }
func (brotliImpl) NewDecoder(r io.Reader) (io.ReadCloser, error) {
	return io.NopCloser(brotli.NewReader(r)), nil
}
func (c brotliImpl) NewEncoder(w io.Writer) (io.WriteCloser, error) {
	return brotli.NewWriterOptions(w, brotli.WriterOptions{}), nil
}

func newDeflate() Compressor                                      { return &deflateImpl{} }
func (deflateImpl) Name() string                                  { return "deflate" }
func (deflateImpl) NewDecoder(r io.Reader) (io.ReadCloser, error) { return flate.NewReader(r), nil }
func (deflateImpl) NewEncoder(w io.Writer) (io.WriteCloser, error) {
	return flate.NewWriter(w, flate.DefaultCompression)
}

func newMediaType(name, problem string, m MarshalFunc, u UnmarshalFunc) MediaType {
	if problem == "" {
		problem = name
	}

	return &mediaTypeImpl{
		normalName:  name,
		problemName: problem,
		marshal:     m,
		unmarshal:   u,
	}
}

func (m *mediaTypeImpl) Name() string                                { return m.normalName }
func (m *mediaTypeImpl) ProblemName() string                         { return m.problemName }
func (m *mediaTypeImpl) Marshal(ctx *Context, v any) ([]byte, error) { return m.marshal(ctx, v) }
func (m *mediaTypeImpl) Unmarshal(r io.Reader, v any) error          { return m.unmarshal(r, v) }

func marshalTest(_ *Context, v any) ([]byte, error) {
	switch vv := v.(type) {
	case error:
		return nil, vv
	default:
		return nil, errors.ErrUnsupported // 测试环境，直接返回 errors.ErrUnsupported 了
	}
}

func unmarshalTest(io.Reader, any) error { return errors.ErrUnsupported }

func marshalJSON(_ *Context, v any) ([]byte, error) { return json.Marshal(v) }

func unmarshalJSON(r io.Reader, v any) error { return json.UnmarshalRead(r, v) }

func marshalXML(_ *Context, v any) ([]byte, error) { return xml.Marshal(v) }

func unmarshalXML(r io.Reader, v any) error { return xml.NewDecoder(r).Decode(v) }

func newCodec(a *assert.Assertion) *Codec {
	c := NewCodec()
	a.NotNil(c)

	c.AddCompressor(newGzip()).
		AddCompressor(newDeflate()).
		//AddCompressor(nil).
		AddMediaType(newMediaType(header.JSON, "application/problem+json", marshalJSON, unmarshalJSON), true, true).
		AddMediaType(newMediaType(header.XML, "application/problem+xml", marshalXML, unmarshalXML), true, true).
		AddMediaType(newMediaType("application/test", "application/problem+test", marshalTest, unmarshalTest), true, true)

	return c
}

func TestCodec_AddMediaType(t *testing.T) {
	a := assert.New(t, false)
	c := NewCodec()

	a.NotPanic(func() {
		c.AddMediaType(newMediaType(header.JSON, "", marshalJSON, unmarshalJSON), true, true)
	})

	a.PanicString(func() {
		c.AddMediaType(newMediaType(header.JSON, "", marshalJSON, unmarshalJSON), true, true)
	}, "存在重复的项 "+header.JSON)

	a.PanicString(func() {
		c.AddMediaType(nil, true, true)
	}, "参数 mt 不能为空")

	mt := newMediaType(header.Plain, "", marshalJSON, unmarshalJSON)
	a.Equal(mt.ProblemName(), header.Plain)
	e := NewCodec().AddMediaType(mt, false, true)
	a.Equal(e.clientAcceptHeader, "").Equal(e.serverAcceptHeader, header.Plain)
}

func TestBuildCompression(t *testing.T) {
	a := assert.New(t, false)

	c := buildCompression(newGzip(), nil)
	a.True(c.wildcard).
		Length(c.types, 0).
		Length(c.wildcardSuffix, 0)

	c = buildCompression(newGzip(), []string{"text"})
	a.Equal(c.types, []string{"text"})

	c = buildCompression(newGzip(), []string{"text", "*"})
	a.Nil(c.types).
		True(c.wildcard).
		Nil(c.wildcardSuffix)
}

func TestCodec_contentEncoding(t *testing.T) {
	a := assert.New(t, false)

	e := NewCodec()
	a.NotNil(e)
	e.AddCompressor(newLZW(), header.Plain, "application/*").
		AddCompressor(newGzip(), header.Plain).
		AddCompressor(newZstd(), "application/*").
		AddCompressor(newBrotli(), "text/*")

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
	e.AddCompressor(newLZW(), header.Plain, "application/*").
		AddCompressor(newGzip(), header.Plain).
		AddCompressor(newGzip(), "application/*")

	a.Equal(e.acceptEncodingHeader, "compress,gzip")

	t.Run("一般", func(t *testing.T) {
		a := a.New(t)

		b, err := e.acceptEncoding(header.JSON, "gzip;q=0.9,br")
		a.NotError(err).NotNil(b).Equal(b.Name(), "gzip")

		// gzip 无法解析，br 能解析但是未添加到 Codec
		b, err = e.acceptEncoding(header.JSON, "gzip;q=x.9,br")
		a.NotError(err).Nil(b)

		b, err = e.acceptEncoding(header.JSON, "br,gzip")
		a.NotError(err).NotNil(b).Equal(b.Name(), "gzip")

		b, err = e.acceptEncoding(header.Plain, "gzip,br")
		a.NotError(err).NotNil(b).Equal(b.Name(), "gzip")

		b, err = e.acceptEncoding(header.Plain, "br")
		a.NotError(err).Nil(b)

		b, err = e.acceptEncoding(header.Plain, "")
		a.NotError(err).Equal(b.Name(), "compress")
	})

	t.Run("header=*", func(t *testing.T) {
		a := a.New(t)

		b, err := e.acceptEncoding(header.XML, "*;q=0")
		a.NotError(err).Nil(b)

		b, err = e.acceptEncoding(header.XML, "*,br")
		a.NotError(err).NotNil(b).Equal(b.Name(), "compress")

		b, err = e.acceptEncoding(header.XML, "*,gzip")
		a.NotError(err).NotNil(b).Equal(b.Name(), "gzip")

		b, err = e.acceptEncoding(header.XML, "*,gzip,compress")
		a.NotError(err).Equal(b.Name(), "gzip")
	})

	t.Run("header=identity", func(t *testing.T) {
		a := assert.New(t, false)
		b, err := e.acceptEncoding(header.XML, "identity,gzip,br")
		a.NotError(err).NotNil(b).Equal(b.Name(), "gzip")

		// 正常匹配
		b, err = e.acceptEncoding(header.XML, "identity;q=0,gzip,br")
		a.NotError(err).NotNil(b).Equal(b.Name(), "gzip")

		// identity=0
		b, err = e.acceptEncoding(header.XML, "identity;q=0,abc,def")
		a.Equal(err, localeutil.Error("client disabled identity")).Nil(b)
	})
}

func TestCodec_contentType(t *testing.T) {
	a := assert.New(t, false)

	mt := NewCodec()
	a.NotNil(mt)
	mt.AddMediaType(newMediaType(header.OctetStream, "", marshalJSON, unmarshalJSON), true, true)

	f, e, err := mt.contentType(";;;")
	a.Error(err).Nil(f).Nil(e)

	// 不存在的 mimetype
	f, e, err = mt.contentType("not-exists; charset=utf-8")
	a.Equal(err, NewError("not found serialization function for %s", "not-exists")).Nil(f).Nil(e)

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

	item, err := mt.accept(header.JSON)
	a.NotError(err).Nil(item)

	item, err = mt.accept("")
	a.NotError(err).Nil(item)

	mt = NewCodec()
	a.NotNil(mt)
	mt.AddMediaType(newMediaType(header.JSON, "", marshalJSON, unmarshalJSON), true, true).
		AddMediaType(newMediaType(header.Plain, "text/plain+problem", marshalXML, unmarshalXML), true, true)

	item, err = mt.accept(header.JSON)
	a.NotError(err).NotNil(item).
		NotNil(item.Marshal).
		Equal(item.Name(), header.JSON).
		Equal(item.ProblemName(), header.JSON)

	// */*
	item, err = mt.accept("*/*")
	a.NotError(err).NotNil(item).
		NotNil(item.Marshal).
		Equal(item.Name(), header.JSON)

	// 空参数，结果同 */*
	item, err = mt.accept("")
	a.NotError(err).NotNil(item).
		NotNil(item.Marshal).
		Equal(item.Name(), header.JSON)

	item, err = mt.accept("*/*,text/plain")
	a.NotError(err).NotNil(item).
		NotNil(item.Marshal).
		Equal(item.Name(), header.Plain).
		Equal(item.ProblemName(), "text/plain+problem")

	// json 被禁，但是匹配 */* 匹配了 text/plain
	item, err = mt.accept(header.JSON + ";q=0,*/*;q=1")
	a.NotError(err).Equal(item.Name(), header.Plain)

	// json 被禁
	item, err = mt.accept(header.JSON + ";q=0")
	a.NotError(err).Nil(item)

	// json 和 text/* 被禁，找不到匹配项
	item, err = mt.accept("text/*;q=0,application/json;q=0,*/*;q=1")
	a.NotError(err).Nil(item)

	item, err = mt.accept("font/wottf;q=x.9")
	a.Error(err).Nil(item)

	item, err = mt.accept("font/wottf")
	a.NotError(err).Nil(item)
}

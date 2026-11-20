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
	"net/http"
	"slices"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/issue9/assert/v6"
	"github.com/issue9/config"
	"github.com/issue9/localeutil"
	"github.com/issue9/mux/v10/header"
	"github.com/klauspost/compress/zstd"
	"golang.org/x/text/language"

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
		method      string
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

func newMediaType(name, problem, method string, m MarshalFunc, u UnmarshalFunc) MediaType {
	if problem == "" {
		problem = name
	}

	return &mediaTypeImpl{
		normalName:  name,
		problemName: problem,
		method:      method,
		marshal:     m,
		unmarshal:   u,
	}
}

func (m *mediaTypeImpl) Name() string                                { return m.normalName }
func (m *mediaTypeImpl) ProblemName() string                         { return m.problemName }
func (m *mediaTypeImpl) Method() string                              { return m.method }
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

func newNegotiation(a *assert.Assertion, conf *config.Config) *Negotiation {
	c := NewNegotiation(language.SimplifiedChinese, conf)
	a.NotNil(c)

	c.AddCompressor(newGzip()).
		AddCompressor(newDeflate()).
		//AddCompressor(nil).
		AddMediaType(newMediaType(header.JSON, "application/problem+json", "", marshalJSON, unmarshalJSON)).
		AddMediaType(newMediaType(header.XML, "application/problem+xml", "", marshalXML, unmarshalXML)).
		AddMediaType(newMediaType("application/test", "application/problem+test", "", marshalTest, unmarshalTest))

	a.Length(slices.Collect(c.MediaTypes()), 3).
		StrictEqual(conf, c.config()).
		StrictEqual(c.Language(), language.SimplifiedChinese)

	return c
}

func TestNegotiation_AddMediaType(t *testing.T) {
	a := assert.New(t, false)
	c := NewNegotiation(language.SimplifiedChinese, nil)

	a.NotPanic(func() {
		c.AddMediaType(newMediaType(header.JSON, "", "", marshalJSON, unmarshalJSON))
	})

	a.PanicString(func() {
		c.AddMediaType(newMediaType(header.JSON, "", "", marshalJSON, unmarshalJSON))
	}, "存在重复的项 "+header.JSON)

	a.PanicString(func() {
		c.AddMediaType()
	}, "参数 mt 不能为空")

	a.PanicString(func() {
		c.AddMediaType(nil)
	}, "参数 mt 不能为 nil")

	t.Run("acceptHeader", func(t *testing.T) {
		mt := newMediaType(header.Plain, "", "", marshalJSON, unmarshalJSON)
		e := NewNegotiation(language.SimplifiedChinese, nil).AddMediaType(mt)
		a.Equal(e.acceptHeader, header.Plain).
			Empty(e.acceptPatchHeader).
			Empty(e.acceptQueryHeader)
	})

	t.Run("acceptPatchHeader", func(t *testing.T) {
		mt := newMediaType(header.Plain, "", http.MethodPatch, marshalJSON, unmarshalJSON)
		e := NewNegotiation(language.SimplifiedChinese, nil).AddMediaType(mt)
		a.Equal(e.acceptPatchHeader, header.Plain).
			Empty(e.acceptHeader).
			Empty(e.acceptQueryHeader)
	})

	t.Run("acceptQueryHeader", func(t *testing.T) {
		mt := newMediaType(header.Plain, "", "QUERY", marshalJSON, unmarshalJSON)
		e := NewNegotiation(language.SimplifiedChinese, nil).AddMediaType(mt)
		a.Equal(e.acceptQueryHeader, header.Plain).
			Empty(e.acceptHeader).
			Empty(e.acceptPatchHeader)
	})

	t.Run("accept", func(t *testing.T) {
		mt1 := newMediaType("application/sql", "", "QUERY", marshalJSON, unmarshalJSON)
		mt2 := newMediaType("text/plain", "", "", marshalJSON, unmarshalJSON)
		mt3 := newMediaType("application/json-patch+json", "", http.MethodPatch, marshalJSON, unmarshalJSON)
		mt4 := newMediaType("application/json-merge+json", "", http.MethodPatch, marshalJSON, unmarshalJSON)
		mt5 := newMediaType("application/cbor", "", "", marshalJSON, unmarshalJSON)
		e := NewNegotiation(language.SimplifiedChinese, nil).AddMediaType(mt1, mt2, mt3, mt4, mt5)
		a.Equal(e.acceptQueryHeader, "application/sql").
			Equal(e.acceptHeader, "text/plain, application/cbor").
			Equal(e.acceptPatchHeader, "application/json-patch+json, application/json-merge+json")
	})
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

func TestNegotiation_contentEncoding(t *testing.T) {
	a := assert.New(t, false)

	e := NewNegotiation(language.SimplifiedChinese, nil)
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

func TestNegotiation_acceptEncoding(t *testing.T) {
	a := assert.New(t, false)

	e := NewNegotiation(language.SimplifiedChinese, nil)
	a.NotNil(e)
	e.AddCompressor(newLZW(), header.Plain, "application/*").
		AddCompressor(newGzip(), header.Plain).
		AddCompressor(newGzip(), "application/*")

	a.Equal(e.acceptEncodingHeader, "compress,gzip")

	t.Run("一般", func(t *testing.T) {
		a := a.New(t)

		b, err := e.acceptEncoding(header.JSON, "gzip;q=0.9,br")
		a.NotError(err).NotNil(b).Equal(b.Name(), "gzip")

		// gzip 无法解析，br 能解析但是未添加到 Negotiation
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

func TestNegotiation_contentType(t *testing.T) {
	a := assert.New(t, false)

	mt := NewNegotiation(language.SimplifiedChinese, nil)
	a.NotNil(mt)
	mt.AddMediaType(
		newMediaType(header.XML, "", "", marshalJSON, unmarshalJSON),
		newMediaType(header.JSON, "", http.MethodPatch, marshalJSON, unmarshalJSON),
		newMediaType(header.CBOR, "", "QUERY", marshalJSON, unmarshalJSON),
	)

	t.Run(";;;", func(t *testing.T) {
		a := a.New(t)
		f, e, err := mt.contentType(";;;", "")
		a.Error(err).Nil(f).Nil(e)
	})

	t.Run("不存在的 mediaType", func(t *testing.T) {
		a := a.New(t)
		f, e, err := mt.contentType("not-exists; charset=utf-8", "")
		a.Equal(err, NewError("not found serialization function for %s", "not-exists")).Nil(f).Nil(e)
	})

	t.Run("xml;charset=utf-8", func(t *testing.T) {
		a := a.New(t)
		f, e, err := mt.contentType(qheader.BuildContentType(header.XML, header.UTF8), "")
		a.NotError(err).NotNil(f).Nil(e)
	})

	t.Run("xml;charset=gb2312", func(t *testing.T) {
		a := a.New(t)
		f, e, err := mt.contentType(qheader.BuildContentType(header.XML, "gb2312"), "")
		a.NotError(err).NotNil(f).NotNil(e)
	})

	t.Run("xml;charset=not-exists", func(t *testing.T) {
		a := a.New(t)
		f, e, err := mt.contentType(qheader.BuildContentType(header.XML, "not-exists"), "")
		a.Error(err).Nil(f).Nil(e)
	})

	t.Run("xml;charset=UTF-8;params", func(t *testing.T) {
		a := a.New(t)
		f, e, err := mt.contentType("application/xml; charset=UTF-8;p1=k1;p2=k2", "")
		a.NotError(err).NotNil(f).Nil(e)
	})

	t.Run("xml;charset=", func(t *testing.T) {
		a := a.New(t)
		f, e, err := mt.contentType(qheader.BuildContentType(header.XML, ""), "")
		a.NotError(err).NotNil(f).Nil(e)
	})

	t.Run("没有 charset", func(t *testing.T) {
		a := a.New(t)
		f, e, err := mt.contentType("application/xml;", "")
		a.NotError(err).NotNil(f).Nil(e)
	})

	t.Run("无 ;charset", func(t *testing.T) {
		a := a.New(t)
		f, e, err := mt.contentType(header.XML, "")
		a.NotError(err).NotNil(f).Nil(e)
	})

	// Patch 在没有匹配 PATCH 的请求的情况下，也会匹配普通项
	t.Run("xml;charset=utf-8;method=PATCH", func(t *testing.T) {
		a := a.New(t)
		f, e, err := mt.contentType(qheader.BuildContentType(header.XML, header.UTF8), http.MethodPatch)
		a.NotError(err).NotNil(f).Nil(e)
	})

	// Patch 匹配 PATCH 项
	t.Run("json;charset=utf-8;method=PATCH", func(t *testing.T) {
		a := a.New(t)
		f, e, err := mt.contentType(qheader.BuildContentType(header.JSON, header.UTF8), http.MethodPatch)
		a.NotError(err).NotNil(f).Nil(e)
	})

	// 普通请求无法匹配 Patch 项
	t.Run("json;charset=utf-8;method=GET", func(t *testing.T) {
		a := a.New(t)
		f, e, err := mt.contentType(qheader.BuildContentType(header.JSON, header.UTF8), http.MethodGet)
		a.Error(err).Nil(f).Nil(e)
	})

	// 普通请求匹配普通项
	t.Run("xml;charset=utf-8;method=GET", func(t *testing.T) {
		a := a.New(t)
		f, e, err := mt.contentType(qheader.BuildContentType(header.XML, header.UTF8), http.MethodGet)
		a.NotError(err).NotNil(f).Nil(e)
	})
}

func TestNegotiation_accept(t *testing.T) {
	a := assert.New(t, false)
	mt := NewNegotiation(language.SimplifiedChinese, nil)
	a.NotNil(mt)

	item, err := mt.accept(header.JSON)
	a.NotError(err).Nil(item)

	item, err = mt.accept("")
	a.NotError(err).Nil(item)

	mt = NewNegotiation(language.SimplifiedChinese, nil)
	a.NotNil(mt)
	mt.AddMediaType(newMediaType(header.JSON, "", "", marshalJSON, unmarshalJSON)).
		AddMediaType(newMediaType(header.Plain, "text/plain+problem", "", marshalXML, unmarshalXML))

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

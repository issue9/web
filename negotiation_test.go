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

	codecImpl struct {
		m MarshalFunc
		u UnmarshalFunc
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

func newCodec(m MarshalFunc, u UnmarshalFunc) Codec              { return &codecImpl{m: m, u: u} }
func (m *codecImpl) Marshal(ctx *Context, v any) ([]byte, error) { return m.m(ctx, v) }
func (m *codecImpl) Unmarshal(r io.Reader, v any) error          { return m.u(r, v) }

func marshalTest(_ *Context, v any) ([]byte, error) {
	switch vv := v.(type) {
	case error:
		return nil, vv
	default:
		return nil, errors.ErrUnsupported // 测试环境，直接返回 errors.ErrUnsupported 了
	}
}

func unmarshalTest(io.Reader, any) error            { return errors.ErrUnsupported }
func marshalJSON(_ *Context, v any) ([]byte, error) { return json.Marshal(v) }
func unmarshalJSON(r io.Reader, v any) error        { return json.UnmarshalRead(r, v) }
func marshalXML(_ *Context, v any) ([]byte, error)  { return xml.Marshal(v) }
func unmarshalXML(r io.Reader, v any) error         { return xml.NewDecoder(r).Decode(v) }

func newNegotiation(a *assert.Assertion, conf *config.Config) *Negotiation {
	c := NewNegotiation(language.SimplifiedChinese, conf)
	a.NotNil(c)

	c.AddCompressor(newGzip()).
		AddCompressor(newDeflate()).
		//AddCompressor(nil).
		AddCodec(newCodec(marshalJSON, unmarshalJSON), header.JSON, "application/problem+json", "", PatchKindNone).
		AddCodec(newCodec(marshalXML, unmarshalXML), header.XML, "application/problem+xml", "", PatchKindNone).
		AddCodec(newCodec(marshalTest, unmarshalTest), "application/test", "application/problem+test", "", PatchKindNone)

	a.Length(slices.Collect(c.MediaTypes()), 3).
		StrictEqual(conf, c.config()).
		StrictEqual(c.Language(), language.SimplifiedChinese)

	return c
}

func TestNewMediaType(t *testing.T) {
	a := assert.New(t, false)

	a.PanicString(func() { NewMediaType(nil, "", "", "", PatchKindNone) }, "参数 codec 不能为 nil")

	a.PanicString(func() {
		NewMediaType(newCodec(marshalJSON, unmarshalJSON), "", "", "", PatchKindNone)
	}, "参数 name 不能为空")

	codec := newCodec(marshalJSON, unmarshalJSON)
	mt := NewPatcherMediaType(codec, "name", "problem")
	a.Equal(mt.Codec, codec).
		StrictEqual(mt.Method(), http.MethodPatch).
		StrictEqual(mt.Name(), "name").
		StrictEqual(mt.ProblemName(), "problem").
		StrictEqual(mt.Method(), http.MethodPatch).
		StrictEqual(mt.PatchKind(), PatchKindPatcher)
}

func TestNegotiation_AddCodec(t *testing.T) {
	a := assert.New(t, false)
	nego := NewNegotiation(language.SimplifiedChinese, nil)

	a.NotPanic(func() { nego.AddCodec(newCodec(marshalJSON, unmarshalJSON), header.JSON, "", "", PatchKindNone) })

	a.PanicString(func() {
		nego.AddCodec(newCodec(marshalJSON, unmarshalJSON), header.JSON, "", "", PatchKindNone)
	}, "存在重复的项 "+header.JSON)

	t.Run("acceptHeader", func(t *testing.T) {
		codec := newCodec(marshalJSON, unmarshalJSON)
		e := NewNegotiation(language.SimplifiedChinese, nil).AddCodec(codec, header.Plain, "", "", PatchKindNone)
		a.Equal(e.acceptHeader, header.Plain).
			Empty(e.acceptPatchHeader).
			Empty(e.acceptQueryHeader)
	})

	t.Run("acceptPatchHeader", func(t *testing.T) {
		codec := newCodec(marshalJSON, unmarshalJSON)
		e := NewNegotiation(language.SimplifiedChinese, nil).AddCodec(codec, header.Plain, "", http.MethodPatch, PatchKindNone)
		a.Equal(e.acceptPatchHeader, header.Plain).
			Empty(e.acceptHeader).
			Empty(e.acceptQueryHeader)
	})

	t.Run("acceptQueryHeader", func(t *testing.T) {
		codec := newCodec(marshalJSON, unmarshalJSON)
		e := NewNegotiation(language.SimplifiedChinese, nil).AddCodec(codec, header.Plain, "", "QUERY", PatchKindNone)
		a.Equal(e.acceptQueryHeader, header.Plain).
			Empty(e.acceptHeader).
			Empty(e.acceptPatchHeader)
	})

	t.Run("accept", func(t *testing.T) {
		mt1 := NewMediaType(newCodec(marshalJSON, unmarshalJSON), "application/sql", "", "QUERY", PatchKindNone)
		mt2 := NewMediaType(newCodec(marshalJSON, unmarshalJSON), "text/plain", "", "", PatchKindNone)
		mt3 := NewMediaType(newCodec(marshalJSON, unmarshalJSON), "application/json-patch+json", "", http.MethodPatch, PatchKindNone)
		mt4 := NewMediaType(newCodec(marshalJSON, unmarshalJSON), "application/json-merge+json", "", http.MethodPatch, PatchKindNone)
		mt5 := NewMediaType(newCodec(marshalJSON, unmarshalJSON), "application/cbor", "", "", PatchKindNone)
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

	nego := NewNegotiation(language.SimplifiedChinese, nil)
	a.NotNil(nego)
	nego.AddCompressor(newLZW(), header.Plain, "application/*").
		AddCompressor(newGzip(), header.Plain).
		AddCompressor(newZstd(), "application/*").
		AddCompressor(newBrotli(), "text/*")

	r := &bytes.Buffer{}
	rc, err := nego.contentEncoding("zstd", r)
	a.NotError(err).NotNil(rc)

	r = bytes.NewBufferString("123")
	rc, err = nego.contentEncoding("", r)
	a.NotError(err).NotNil(rc)
	data, err := io.ReadAll(rc)
	a.NotError(err).Equal(string(data), "123")
}

func TestNegotiation_acceptEncoding(t *testing.T) {
	a := assert.New(t, false)

	nego := NewNegotiation(language.SimplifiedChinese, nil)
	a.NotNil(nego)
	nego.AddCompressor(newLZW(), header.Plain, "application/*").
		AddCompressor(newGzip(), header.Plain).
		AddCompressor(newGzip(), "application/*")

	a.Equal(nego.acceptEncodingHeader, "compress,gzip")

	t.Run("一般", func(t *testing.T) {
		a := a.New(t)

		b, err := nego.acceptEncoding(header.JSON, "gzip;q=0.9,br")
		a.NotError(err).NotNil(b).Equal(b.Name(), "gzip")

		// gzip 无法解析，br 能解析但是未添加到 Negotiation
		b, err = nego.acceptEncoding(header.JSON, "gzip;q=x.9,br")
		a.NotError(err).Nil(b)

		b, err = nego.acceptEncoding(header.JSON, "br,gzip")
		a.NotError(err).NotNil(b).Equal(b.Name(), "gzip")

		b, err = nego.acceptEncoding(header.Plain, "gzip,br")
		a.NotError(err).NotNil(b).Equal(b.Name(), "gzip")

		b, err = nego.acceptEncoding(header.Plain, "br")
		a.NotError(err).Nil(b)

		b, err = nego.acceptEncoding(header.Plain, "")
		a.NotError(err).Equal(b.Name(), "compress")
	})

	t.Run("header=*", func(t *testing.T) {
		a := a.New(t)

		b, err := nego.acceptEncoding(header.XML, "*;q=0")
		a.NotError(err).Nil(b)

		b, err = nego.acceptEncoding(header.XML, "*,br")
		a.NotError(err).NotNil(b).Equal(b.Name(), "compress")

		b, err = nego.acceptEncoding(header.XML, "*,gzip")
		a.NotError(err).NotNil(b).Equal(b.Name(), "gzip")

		b, err = nego.acceptEncoding(header.XML, "*,gzip,compress")
		a.NotError(err).Equal(b.Name(), "gzip")
	})

	t.Run("header=identity", func(t *testing.T) {
		a := assert.New(t, false)
		b, err := nego.acceptEncoding(header.XML, "identity,gzip,br")
		a.NotError(err).NotNil(b).Equal(b.Name(), "gzip")

		// 正常匹配
		b, err = nego.acceptEncoding(header.XML, "identity;q=0,gzip,br")
		a.NotError(err).NotNil(b).Equal(b.Name(), "gzip")

		// identity=0
		b, err = nego.acceptEncoding(header.XML, "identity;q=0,abc,def")
		a.Equal(err, localeutil.Error("client disabled identity")).Nil(b)
	})
}

func TestNegotiation_contentType(t *testing.T) {
	a := assert.New(t, false)

	nego := NewNegotiation(language.SimplifiedChinese, nil)
	a.NotNil(nego)
	nego.AddMediaType(
		NewMediaType(newCodec(marshalJSON, unmarshalJSON), header.XML, "", "", PatchKindNone),
		NewMediaType(newCodec(marshalJSON, unmarshalJSON), header.JSON, "", http.MethodPatch, PatchKindNone),
		NewMediaType(newCodec(marshalJSON, unmarshalJSON), header.CBOR, "", "QUERY", PatchKindNone),
	)

	t.Run(";;;", func(t *testing.T) {
		a := a.New(t)
		f, e, err := nego.contentType(";;;", "")
		a.Error(err).Nil(f).Nil(e)
	})

	t.Run("不存在的 mediaType", func(t *testing.T) {
		a := a.New(t)
		f, e, err := nego.contentType("not-exists; charset=utf-8", "")
		a.Equal(err, NewError("not found serialization function for %s", "not-exists")).Nil(f).Nil(e)
	})

	t.Run("xml;charset=utf-8", func(t *testing.T) {
		a := a.New(t)
		f, e, err := nego.contentType(qheader.BuildContentType(header.XML, header.UTF8), "")
		a.NotError(err).NotNil(f).Nil(e)
	})

	t.Run("xml;charset=gb2312", func(t *testing.T) {
		a := a.New(t)
		f, e, err := nego.contentType(qheader.BuildContentType(header.XML, "gb2312"), "")
		a.NotError(err).NotNil(f).NotNil(e)
	})

	t.Run("xml;charset=not-exists", func(t *testing.T) {
		a := a.New(t)
		f, e, err := nego.contentType(qheader.BuildContentType(header.XML, "not-exists"), "")
		a.Error(err).Nil(f).Nil(e)
	})

	t.Run("xml;charset=UTF-8;params", func(t *testing.T) {
		a := a.New(t)
		f, e, err := nego.contentType("application/xml; charset=UTF-8;p1=k1;p2=k2", "")
		a.NotError(err).NotNil(f).Nil(e)
	})

	t.Run("xml;charset=", func(t *testing.T) {
		a := a.New(t)
		f, e, err := nego.contentType(qheader.BuildContentType(header.XML, ""), "")
		a.NotError(err).NotNil(f).Nil(e)
	})

	t.Run("没有 charset", func(t *testing.T) {
		a := a.New(t)
		f, e, err := nego.contentType("application/xml;", "")
		a.NotError(err).NotNil(f).Nil(e)
	})

	t.Run("无 ;charset", func(t *testing.T) {
		a := a.New(t)
		f, e, err := nego.contentType(header.XML, "")
		a.NotError(err).NotNil(f).Nil(e)
	})

	// Patch 在没有匹配 PATCH 的请求的情况下，也会匹配普通项
	t.Run("xml;charset=utf-8;method=PATCH", func(t *testing.T) {
		a := a.New(t)
		f, e, err := nego.contentType(qheader.BuildContentType(header.XML, header.UTF8), http.MethodPatch)
		a.NotError(err).NotNil(f).Nil(e)
	})

	// Patch 匹配 PATCH 项
	t.Run("json;charset=utf-8;method=PATCH", func(t *testing.T) {
		a := a.New(t)
		f, e, err := nego.contentType(qheader.BuildContentType(header.JSON, header.UTF8), http.MethodPatch)
		a.NotError(err).NotNil(f).Nil(e)
	})

	// 普通请求无法匹配 Patch 项
	t.Run("json;charset=utf-8;method=GET", func(t *testing.T) {
		a := a.New(t)
		f, e, err := nego.contentType(qheader.BuildContentType(header.JSON, header.UTF8), http.MethodGet)
		a.Error(err).Nil(f).Nil(e)
	})

	// 普通请求匹配普通项
	t.Run("xml;charset=utf-8;method=GET", func(t *testing.T) {
		a := a.New(t)
		f, e, err := nego.contentType(qheader.BuildContentType(header.XML, header.UTF8), http.MethodGet)
		a.NotError(err).NotNil(f).Nil(e)
	})
}

func TestNegotiation_accept(t *testing.T) {
	a := assert.New(t, false)
	nego := NewNegotiation(language.SimplifiedChinese, nil)
	a.NotNil(nego)

	item, err := nego.accept(header.JSON)
	a.NotError(err).Nil(item)

	item, err = nego.accept("")
	a.NotError(err).Nil(item)

	nego = NewNegotiation(language.SimplifiedChinese, nil)
	a.NotNil(nego)
	nego.AddCodec(newCodec(marshalJSON, unmarshalJSON), header.JSON, "", "", PatchKindNone).
		AddCodec(newCodec(marshalXML, unmarshalXML), header.Plain, "text/plain+problem", "", PatchKindNone).
		AddCodec(newCodec(marshalXML, unmarshalXML), header.XML, "", http.MethodPatch, PatchKindNone)

	t.Run("JSON", func(t *testing.T) {
		a := a.New(t)
		item, err = nego.accept(header.JSON)
		a.NotError(err).NotNil(item).
			NotNil(item.Marshal).
			StrictEqual(item.Name(), header.JSON).
			StrictEqual(item.ProblemName(), header.JSON)
	})

	t.Run("*/*", func(t *testing.T) {
		a := a.New(t)
		item, err = nego.accept("*/*")
		a.NotError(err).NotNil(item).
			NotNil(item.Marshal).
			StrictEqual(item.Name(), header.JSON)
	})

	t.Run("空参数，结果同 */*", func(t *testing.T) {
		a := a.New(t)
		item, err = nego.accept("")
		a.NotError(err).NotNil(item).
			NotNil(item.Marshal).
			StrictEqual(item.Name(), header.JSON)
	})

	t.Run("*/*,text/plain", func(t *testing.T) {
		a := a.New(t)
		item, err = nego.accept("*/*,text/plain")
		a.NotError(err).NotNil(item).
			NotNil(item.Marshal).
			StrictEqual(item.Name(), header.Plain).
			StrictEqual(item.ProblemName(), "text/plain+problem")
	})

	t.Run("json 被禁，但是 */* 匹配了 text/plain", func(t *testing.T) {
		a := a.New(t)
		item, err = nego.accept(header.JSON + ";q=0,*/*;q=1")
		a.NotError(err).StrictEqual(item.Name(), header.Plain)
	})

	t.Run("json 被禁", func(t *testing.T) {
		a := a.New(t)
		item, err = nego.accept(header.JSON + ";q=0")
		a.NotError(err).Nil(item)
	})

	t.Run("json 和 text/* 被禁，XML 可用", func(t *testing.T) {
		a := a.New(t)
		item, err = nego.accept("text/*;q=0,application/json;q=0,*/*;q=1")
		a.NotError(err).NotNil(item).StrictEqual(item.Name(), header.XML)
	})

	t.Run("font/wottf;q=x.9", func(t *testing.T) {
		a := a.New(t)
		item, err = nego.accept("font/wottf;q=x.9")
		a.Error(err).Nil(item)
	})

	t.Run("font/wottf", func(t *testing.T) {
		a := a.New(t)
		item, err = nego.accept("font/wottf")
		a.NotError(err).Nil(item)
	})
}

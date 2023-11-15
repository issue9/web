// SPDX-License-Identifier: MIT

package web

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/lzw"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/localeutil"

	"github.com/issue9/web/codec/compressor"
	"github.com/issue9/web/locales"
)

const testMimetype = "application/octet-stream"

var testMimetypes = []*Mimetype{
	{
		Name:      "application/json",
		Problem:   "application/problem+json",
		Marshal:   func(_ *Context, v any) ([]byte, error) { return json.Marshal(v) },
		Unmarshal: func(r io.Reader, v any) error { return json.NewDecoder(r).Decode(v) },
	},
	{
		Name:      "application/xml",
		Problem:   "application/problem+xml",
		Marshal:   func(_ *Context, v any) ([]byte, error) { return xml.Marshal(v) },
		Unmarshal: func(r io.Reader, v any) error { return xml.NewDecoder(r).Decode(v) },
	},
	{
		Name:      "application/test",
		Problem:   "application/problem+test",
		Marshal:   marshalTest,
		Unmarshal: unmarshalTest,
	},
	{Name: "nil", Problem: "nil"},
}

type compressorTest struct {
	name string
}

func (c *compressorTest) Name() string { return c.name }

func (c *compressorTest) NewEncoder(w io.Writer) (io.WriteCloser, error) {
	switch c.name {
	case "gzip":
		return gzip.NewWriter(w), nil
	case "deflate":
		return flate.NewWriter(w, 8)
	default:
		return nil, nil
	}
}

func (c *compressorTest) NewDecoder(r io.Reader) (io.ReadCloser, error) {
	switch c.name {
	case "gzip":
		return gzip.NewReader(r)
	case "deflate":
		return flate.NewReader(r), nil
	default:
		if c.name != "" {
			return nil, fmt.Errorf("不支持的压缩方法 %s", c.name)
		}
		return io.NopCloser(r), nil
	}
}

func marshalTest(_ *Context, v any) ([]byte, error) {
	switch vv := v.(type) {
	case error:
		return nil, vv
	default:
		return nil, ErrUnsupportedSerialization()
	}
}

func unmarshalTest(r io.Reader, v any) error {
	return ErrUnsupportedSerialization()
}

func newCodec(a *assert.Assertion) *Codec {
	ms := []*Mimetype{
		{
			Name:      "application/json",
			Problem:   "application/problem+json",
			Marshal:   func(_ *Context, v any) ([]byte, error) { return json.Marshal(v) },
			Unmarshal: func(r io.Reader, v any) error { return json.NewDecoder(r).Decode(v) },
		},
		{
			Name:      "application/xml",
			Problem:   "application/problem+xml",
			Marshal:   func(_ *Context, v any) ([]byte, error) { return xml.Marshal(v) },
			Unmarshal: func(r io.Reader, v any) error { return xml.NewDecoder(r).Decode(v) },
		},
		{
			Name:      "application/test",
			Problem:   "application/problem+test",
			Marshal:   marshalTest,
			Unmarshal: unmarshalTest,
		},
		{Name: "nil", Problem: "nil"},
	}

	cs := []*Compression{
		{Compressor: &compressorTest{name: "gzip"}},
		{Compressor: &compressorTest{name: "deflate"}},
		{Compressor: &compressorTest{name: ""}},
	}

	c, err := NewCodec("ms", "cs", ms, cs)
	a.NotError(err).NotNil(c)
	return c
}

func TestNewCodec(t *testing.T) {
	a := assert.New(t, false)

	c, err := NewCodec("", "", nil, nil)
	a.NotError(err).NotNil(c)

	c, err = NewCodec("ms", "cs", []*Mimetype{
		{Name: "application/json", Marshal: func(*Context, any) ([]byte, error) { return nil, nil }},
	}, nil)
	a.NotError(err).NotNil(c)

	c, err = NewCodec("ms", "cs", []*Mimetype{
		{Name: "application/json", Marshal: func(*Context, any) ([]byte, error) { return nil, nil }},
		{Name: "application/json", Marshal: func(*Context, any) ([]byte, error) { return nil, nil }},
	}, nil)
	a.Equal(err.Message, locales.DuplicateValue).Nil(c).
		Equal(err.Field, "ms[0].Name")

	c, err = NewCodec("ms", "cs", []*Mimetype{
		{Name: "", Marshal: func(*Context, any) ([]byte, error) { return nil, nil }},
	}, nil)
	a.Equal(err.Field, "ms[0].Name").Nil(c)

	c, err = NewCodec("ms", "cs", nil, []*Compression{{}})
	a.Equal(err.Field, "cs[0].Compressor").Nil(c)
}

func TestMimetype_sanitize(t *testing.T) {
	a := assert.New(t, false)

	m := &Mimetype{}
	err := m.sanitize()
	a.Error(err).Equal(err.Field, "Name")

	m = &Mimetype{Name: "test"}
	err = m.sanitize()
	a.NotError(err).Equal(m.Problem, m.Name)

	m = &Mimetype{Name: "test", Problem: "p"}
	err = m.sanitize()
	a.NotError(err).
		Equal(m.Problem, "p").
		Equal(m.Name, "test")
}

func TestCompression_sanitize(t *testing.T) {
	a := assert.New(t, false)

	c := &Compression{}
	err := c.sanitize()
	a.Error(err).Equal(err.Field, "Compressor")

	c = &Compression{Compressor: compressor.NewZstdCompressor()}
	err = c.sanitize()
	a.NotError(err).
		True(c.wildcard).
		Length(c.Types, 0).
		Length(c.wildcardSuffix, 0)

	c = &Compression{Compressor: compressor.NewZstdCompressor(), Types: []string{"text"}}
	err = c.sanitize()
	a.NotError(err).Equal(c.Types, []string{"text"})
}

func TestCodec_ContentEncoding(t *testing.T) {
	a := assert.New(t, false)

	e, fe := NewCodec("ms", "cs", nil, []*Compression{
		{Compressor: compressor.NewLZWCompressor(lzw.LSB, 2), Types: []string{"text/plain", "application/*"}},
		{Compressor: compressor.NewGzipCompressor(3), Types: []string{"text/plain"}},
		{Compressor: compressor.NewZstdCompressor(), Types: []string{"application/*"}},
	})
	a.NotError(fe).NotNil(e)

	r := &bytes.Buffer{}
	rc, err := e.contentEncoding("zstd", r)
	a.NotError(err).NotNil(rc)

	r = bytes.NewBufferString("123")
	rc, err = e.contentEncoding("", r)
	a.NotError(err).NotNil(rc)
	data, err := io.ReadAll(rc)
	a.NotError(err).Equal(string(data), "123")
}

func TestCodec_AcceptEncoding(t *testing.T) {
	a := assert.New(t, false)

	e, err := NewCodec("ms", "cs", nil, []*Compression{
		{Compressor: compressor.NewLZWCompressor(lzw.LSB, 2), Types: []string{"text/plain", "application/*"}},
		{Compressor: compressor.NewGzipCompressor(3), Types: []string{"text/plain"}},
		{Compressor: compressor.NewGzipCompressor(9), Types: []string{"application/*"}},
	})
	a.NotError(err).NotNil(e)

	a.Equal(e.acceptEncodingHeader, "compress,gzip")

	t.Run("一般", func(t *testing.T) {
		a := assert.New(t, false)
		b, name, notAccept := e.acceptEncoding("application/json", "gzip;q=0.9,br", nil)
		a.False(notAccept).NotNil(b).Equal(name, "gzip")

		b, name, notAccept = e.acceptEncoding("application/json", "br,gzip", nil)
		a.False(notAccept).NotNil(b).Equal(name, "gzip")

		b, name, notAccept = e.acceptEncoding("text/plain", "gzip,br", nil)
		a.False(notAccept).NotNil(b).Equal(name, "gzip")

		b, _, notAccept = e.acceptEncoding("text/plain", "br", nil)
		a.False(notAccept).Nil(b)

		b, _, notAccept = e.acceptEncoding("text/plain", "", nil)
		a.False(notAccept).Nil(b)
	})

	t.Run("header=*", func(t *testing.T) {
		a := assert.New(t, false)
		b, _, notAccept := e.acceptEncoding("application/xml", "*;q=0", nil)
		a.True(notAccept).Nil(b)

		b, name, notAccept := e.acceptEncoding("application/xml", "*,br", nil)
		a.False(notAccept).NotNil(b).Equal(name, "compress")

		b, name, notAccept = e.acceptEncoding("application/xml", "*,gzip", nil)
		a.False(notAccept).NotNil(b).Equal(name, "compress")

		b, _, notAccept = e.acceptEncoding("application/xml", "*,gzip,compress", nil) // gzip,compress 都排除了
		a.False(notAccept).Nil(b)
	})

	t.Run("header=identity", func(t *testing.T) {
		a := assert.New(t, false)
		b, name, notAccept := e.acceptEncoding("application/xml", "identity,gzip,br", nil)
		a.False(notAccept).NotNil(b).Equal(name, "gzip")

		// 正常匹配
		b, name, notAccept = e.acceptEncoding("application/xml", "identity;q=0,gzip,br", nil)
		a.False(notAccept).NotNil(b).Equal(name, "gzip")

		// 没有可匹配，选取第一个
		b, _, notAccept = e.acceptEncoding("application/xml", "identity;q=0,abc,def", nil)
		a.False(notAccept).Nil(b)
	})
}

func TestCodec_ContentType(t *testing.T) {
	a := assert.New(t, false)

	mt, fe := NewCodec("ms", "cs", []*Mimetype{
		{
			Name:      testMimetype,
			Problem:   "",
			Marshal:   func(_ *Context, v any) ([]byte, error) { return json.Marshal(v) },
			Unmarshal: func(r io.Reader, v any) error { return json.NewDecoder(r).Decode(v) },
		},
	}, nil)
	a.NotError(fe).NotNil(mt)

	f, e, err := mt.contentType(";;;")
	a.Error(err).Nil(f).Nil(e)

	// 不存在的 mimetype
	f, e, err = mt.contentType("not-exists; charset=utf-8")
	a.Equal(err, localeutil.Error("not found serialization function for %s", "not-exists")).Nil(f).Nil(e)

	// charset=utf-8
	f, e, err = mt.contentType("application/octet-stream; charset=utf-8")
	a.NotError(err).NotNil(f).Nil(e)

	// charset=gb2312
	f, e, err = mt.contentType("application/octet-stream; charset=gb2312")
	a.NotError(err).NotNil(f).NotNil(e)

	// charset=not-exists
	f, e, err = mt.contentType("application/octet-stream; charset=not-exists")
	a.Error(err).Nil(f).Nil(e)

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

func TestCodec_Accept(t *testing.T) {
	a := assert.New(t, false)
	mt, fe := NewCodec("ms", "cs", nil, nil)
	a.NotError(fe).NotNil(mt)

	item := mt.accept(testMimetype)
	a.Nil(item)

	item = mt.accept("")
	a.Nil(item)

	mt, fe = NewCodec("ms", "cs", []*Mimetype{
		{
			Name:      testMimetype,
			Problem:   "",
			Marshal:   func(_ *Context, v any) ([]byte, error) { return json.Marshal(v) },
			Unmarshal: func(r io.Reader, v any) error { return json.NewDecoder(r).Decode(v) },
		},
		{
			Name:      "text/plain",
			Problem:   "text/plain+problem",
			Marshal:   func(_ *Context, v any) ([]byte, error) { return xml.Marshal(v) },
			Unmarshal: func(r io.Reader, v any) error { return xml.NewDecoder(r).Decode(v) },
		},
		{
			Name:      "empty",
			Marshal:   nil,
			Unmarshal: nil,
			Problem:   "",
		},
	}, nil)
	a.NotError(fe).NotNil(mt)

	item = mt.accept(testMimetype)
	a.NotNil(item).
		NotNil(item.Marshal).
		Equal(item.name(false), testMimetype).
		Equal(item.name(true), testMimetype)

	// */* 如果指定了 DefaultMimetype，则必定是该值
	item = mt.accept("*/*")
	a.NotNil(item).
		NotNil(item.Marshal).
		Equal(item.name(false), testMimetype)

	// 同 */*
	item = mt.accept("")
	a.NotNil(item).
		NotNil(item.Marshal).
		Equal(item.name(false), testMimetype)

	item = mt.accept("*/*,text/plain")
	a.NotNil(item).
		NotNil(item.Marshal).
		Equal(item.name(false), "text/plain").
		Equal(item.name(true), "text/plain+problem")

	item = mt.accept("font/wottf;q=x.9")
	a.Nil(item)

	item = mt.accept("font/wottf")
	a.Nil(item)

	// 匹配 empty
	item = mt.accept("empty")
	a.NotNil(item).
		Equal(item.name(false), "empty").
		Nil(item.Marshal)
}

func TestCodec_findMarshal(t *testing.T) {
	a := assert.New(t, false)
	mm, fe := NewCodec("ms", "cs", []*Mimetype{
		{Name: "text", Marshal: nil, Unmarshal: nil, Problem: ""},
		{Name: "text/plain", Marshal: nil, Unmarshal: nil, Problem: ""},
		{Name: "text/text", Marshal: nil, Unmarshal: nil, Problem: ""},
		{Name: "application/aa", Marshal: nil, Unmarshal: nil, Problem: ""},
		{Name: "application/bb", Marshal: nil, Unmarshal: nil, Problem: "application/problem+bb"},
		{Name: testMimetype, Marshal: nil, Unmarshal: nil, Problem: ""},
	}, nil)
	a.NotError(fe).NotNil(mm)

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

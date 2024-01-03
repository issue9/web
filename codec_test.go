// SPDX-License-Identifier: MIT

package web

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/localeutil"
	"github.com/klauspost/compress/zstd"

	"github.com/issue9/web/mimetype"
)

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
	case "zstd":
		return zstd.NewWriter(w)
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
	case "zstd":
		rr, err := zstd.NewReader(r)
		return io.NopCloser(rr), err
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
		return nil, mimetype.ErrUnsupported()
	}
}

func unmarshalTest(r io.Reader, v any) error { return mimetype.ErrUnsupported() }

func marshalJSON(_ *Context, v any) ([]byte, error) { return json.Marshal(v) }

func unmarshalJSON(r io.Reader, v any) error { return json.NewDecoder(r).Decode(v) }

func marshalXML(_ *Context, v any) ([]byte, error) { return xml.Marshal(v) }

func unmarshalXML(r io.Reader, v any) error { return xml.NewDecoder(r).Decode(v) }

func newCodec(a *assert.Assertion) *Codec {
	c := NewCodec()
	a.NotNil(c)

	c.AddCompressor(&compressorTest{name: "gzip"}).
		AddCompressor(&compressorTest{name: "deflate"}).
		AddCompressor(&compressorTest{name: ""}).
		AddMimetype("application/json", marshalJSON, unmarshalJSON, "application/problem+json").
		AddMimetype("application/xml", marshalXML, unmarshalXML, "application/problem+xml").
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
		c.AddMimetype("application/json", nil, nil, "")
	}, "参数 m 不能为空")

	a.PanicString(func() {
		c.AddMimetype("application/json", marshalJSON, nil, "")
	}, "参数 u 不能为空")

	a.NotPanic(func() {
		c.AddMimetype("application/json", marshalJSON, unmarshalJSON, "")
	})

	a.PanicString(func() {
		c.AddMimetype("application/json", marshalJSON, unmarshalJSON, "")
	}, "存在重复的项 application/json")
}

func TestBuildCompression(t *testing.T) {
	a := assert.New(t, false)

	c := buildCompression(&compressorTest{name: "gzip"}, nil)
	a.True(c.wildcard).
		Length(c.types, 0).
		Length(c.wildcardSuffix, 0)

	c = buildCompression(&compressorTest{name: "gzip"}, []string{"text"})
	a.Equal(c.types, []string{"text"})

	c = buildCompression(&compressorTest{name: "gzip"}, []string{"text", "*"})
	a.Nil(c.types).
		True(c.wildcard).
		Nil(c.wildcardSuffix)
}

func TestCodec_contentEncoding(t *testing.T) {
	a := assert.New(t, false)

	e := NewCodec()
	a.NotNil(e)
	e.AddCompressor(&compressorTest{name: "compress"}, "text/plain", "application/*").
		AddCompressor(&compressorTest{name: "gzip"}, "text/plain").
		AddCompressor(&compressorTest{name: "zstd"}, "application/*")

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
	e.AddCompressor(&compressorTest{name: "compress"}, "text/plain", "application/*").
		AddCompressor(&compressorTest{name: "gzip"}, "text/plain").
		AddCompressor(&compressorTest{name: "gzip"}, "application/*")

	a.Equal(e.acceptEncodingHeader, "compress,gzip")

	t.Run("一般", func(t *testing.T) {
		a := assert.New(t, false)
		b, notAccept := e.acceptEncoding("application/json", "gzip;q=0.9,br")
		a.False(notAccept).NotNil(b).Equal(b.Name(), "gzip")

		b, notAccept = e.acceptEncoding("application/json", "br,gzip")
		a.False(notAccept).NotNil(b).Equal(b.Name(), "gzip")

		b, notAccept = e.acceptEncoding("text/plain", "gzip,br")
		a.False(notAccept).NotNil(b).Equal(b.Name(), "gzip")

		b, notAccept = e.acceptEncoding("text/plain", "br")
		a.False(notAccept).Nil(b)

		b, notAccept = e.acceptEncoding("text/plain", "")
		a.False(notAccept).Nil(b)
	})

	t.Run("header=*", func(t *testing.T) {
		a := assert.New(t, false)
		b, notAccept := e.acceptEncoding("application/xml", "*;q=0")
		a.True(notAccept).Nil(b)

		b, notAccept = e.acceptEncoding("application/xml", "*,br")
		a.False(notAccept).NotNil(b).Equal(b.Name(), "compress")

		b, notAccept = e.acceptEncoding("application/xml", "*,gzip")
		a.False(notAccept).NotNil(b).Equal(b.Name(), "compress")

		b, notAccept = e.acceptEncoding("application/xml", "*,gzip,compress") // gzip,compress 都排除了
		a.False(notAccept).Nil(b)
	})

	t.Run("header=identity", func(t *testing.T) {
		a := assert.New(t, false)
		b, notAccept := e.acceptEncoding("application/xml", "identity,gzip,br")
		a.False(notAccept).NotNil(b).Equal(b.Name(), "gzip")

		// 正常匹配
		b, notAccept = e.acceptEncoding("application/xml", "identity;q=0,gzip,br")
		a.False(notAccept).NotNil(b).Equal(b.Name(), "gzip")

		// 没有可匹配，选取第一个
		b, notAccept = e.acceptEncoding("application/xml", "identity;q=0,abc,def")
		a.False(notAccept).Nil(b)
	})
}

func TestCodec_contentType(t *testing.T) {
	a := assert.New(t, false)

	mt := NewCodec()
	a.NotNil(mt)
	mt.AddMimetype("application/octet-stream", marshalJSON, unmarshalJSON, "")

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

func TestCodec_accept(t *testing.T) {
	a := assert.New(t, false)
	mt := NewCodec()
	a.NotNil(mt)

	item := mt.accept("application/json")
	a.Nil(item)

	item = mt.accept("")
	a.Nil(item)

	mt = NewCodec()
	a.NotNil(mt)
	mt.AddMimetype("application/json", marshalJSON, unmarshalJSON, "").
		AddMimetype("text/plain", marshalXML, unmarshalXML, "text/plain+problem")

	item = mt.accept("application/json")
	a.NotNil(item).
		NotNil(item.Marshal).
		Equal(item.name(false), "application/json").
		Equal(item.name(true), "application/json")

	// */*
	item = mt.accept("*/*")
	a.NotNil(item).
		NotNil(item.Marshal).
		Equal(item.name(false), "application/json")

	// 空参数，结果同 */*
	item = mt.accept("")
	a.NotNil(item).
		NotNil(item.Marshal).
		Equal(item.name(false), "application/json")

	item = mt.accept("*/*,text/plain")
	a.NotNil(item).
		NotNil(item.Marshal).
		Equal(item.name(false), "text/plain").
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
		AddMimetype("application/json", marshalTest, unmarshalTest, "")

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

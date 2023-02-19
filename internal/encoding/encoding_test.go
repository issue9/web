// SPDX-License-Identifier: MIT

package encoding

import (
	"bytes"
	"compress/gzip"
	"compress/lzw"
	"io"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/issue9/assert/v3"
	"github.com/klauspost/compress/zstd"
)

func TestEncodings_Add(t *testing.T) {
	a := assert.New(t, false)

	e := NewEncodings(nil)
	a.NotNil(e)

	e.Add("gzip", GZipWriter(gzip.DefaultCompression), "text/*")
	e.Add("gzip", GZipWriter(gzip.DefaultCompression), "text/*")
	a.Equal(2, len(e.algs))

	a.PanicString(func() {
		e.Add("gzip", nil, "text/*")
	}, "参数 f 不能为空")

	a.PanicString(func() {
		e.Add("*", GZipWriter(gzip.DefaultCompression), "text/*")
	}, "name 值不能为 identity 和 *")

	a.PanicString(func() {
		e.Add("identity", GZipWriter(gzip.DefaultCompression), "text")
	}, "name 值不能为 identity 和 *")
}

func TestEncodings_Search(t *testing.T) {
	a := assert.New(t, false)

	e := NewEncodings(nil)
	a.NotNil(e)

	e.Add("compress", CompressWriter(lzw.LSB, 2), "text/plain", "application/*")
	//e.Add("lzw-msb-2", "compress", CompressWriter(lzw.MSB, 2))
	//e.Add("lzw-msb-5", "compress", CompressWriter(lzw.MSB, 5))
	e.Add("gzip", GZipWriter(3), "text/plain")
	e.Add("gzip", GZipWriter(9), "application/*")
	//e.Add("deflate-9", "deflate", DeflateWriter(9))

	t.Run("一般", func(t *testing.T) {
		a := assert.New(t, false)
		b, notAccept := e.Search("application/json", "gzip;q=0.9,br")
		a.False(notAccept).NotNil(b).Equal(b.name, "gzip")

		b, notAccept = e.Search("application/json", "br,gzip")
		a.False(notAccept).NotNil(b).Equal(b.name, "gzip")

		b, notAccept = e.Search("text/plain", "gzip,br")
		a.False(notAccept).NotNil(b).Equal(b.name, "gzip")

		b, notAccept = e.Search("text/plain", "br")
		a.False(notAccept).Nil(b)

		b, notAccept = e.Search("text/plain", "")
		a.False(notAccept).Nil(b)
	})

	t.Run("header=*", func(t *testing.T) {
		a := assert.New(t, false)
		b, notAccept := e.Search("application/xml", "*;q=0")
		a.True(notAccept).Nil(b)

		b, notAccept = e.Search("application/xml", "*,br")
		a.False(notAccept).NotNil(b).Equal(b.name, "compress")

		b, notAccept = e.Search("application/xml", "*,gzip")
		a.False(notAccept).NotNil(b).Equal(b.name, "compress")

		b, notAccept = e.Search("application/xml", "*,gzip,compress") // gzip,compress 都排除了
		a.False(notAccept).Nil(b)
	})

	t.Run("header=identity", func(t *testing.T) {
		a := assert.New(t, false)
		b, notAccept := e.Search("application/xml", "identity,gzip,br")
		a.False(notAccept).NotNil(b).Equal(b.name, "gzip")

		// 正常匹配
		b, notAccept = e.Search("application/xml", "identity;q=0,gzip,br")
		a.False(notAccept).NotNil(b).Equal(b.name, "gzip")

		// 没有可匹配，选取第一个
		b, notAccept = e.Search("application/xml", "identity;q=0,abc,def")
		a.False(notAccept).Nil(b)
	})
}

func TestEncodings_Compress(t *testing.T) {
	a := assert.New(t, false)

	e := NewEncodings(nil)
	a.NotNil(e)

	e.Add("compress", CompressWriter(lzw.LSB, 2), "text/*", "application/*", "application/xml")
	e.Add("gzip", GZipWriter(3), "text/*", "application/*")
	e.Add("br", BrotliWriter(brotli.WriterOptions{Quality: 3, LGWin: 10}), "application/*")
	e.Add("zstd", ZstdWriter(zstd.WithEncoderLevel(zstd.SpeedDefault)), "application/xml")

	// br

	b, notAccept := e.Search("application/json", "gzip;q=0.9,br")
	a.False(notAccept).NotNil(b).Equal(b.name, "br")

	w := &bytes.Buffer{}
	wc := b.Get(w)
	_, err := wc.Write([]byte("123456"))
	a.NotError(err)
	a.NotError(wc.Close())
	a.NotEqual(w.String(), "123456").NotEmpty(w.String())

	brotliR := brotli.NewReader(w)
	a.NotNil(brotliR)
	data, err := io.ReadAll(brotliR)
	a.NotError(err).NotNil(data).Equal(string(data), "123456")

	// gzip

	b, notAccept = e.Search("application/json", "gzip;q=0.9")
	a.False(notAccept).NotNil(b).Equal(b.name, "gzip")

	w = &bytes.Buffer{}
	wc = b.Get(w)
	_, err = wc.Write([]byte("123456"))
	a.NotError(err)
	a.NotError(wc.Close())
	a.NotEqual(w.String(), "123456").NotEmpty(w.String())

	gzipR, err := gzip.NewReader(w)
	a.NotError(err).NotNil(gzipR)
	data, err = io.ReadAll(gzipR)
	a.NotError(err).NotNil(data).Equal(string(data), "123456")

	// zstd

	b, notAccept = e.Search("application/xml", "zstd")
	a.False(notAccept).NotNil(b).Equal(b.name, "zstd")

	w = &bytes.Buffer{}
	wc = b.Get(w)
	_, err = wc.Write([]byte("123456"))
	a.NotError(err)
	a.NotError(wc.Close())
	a.NotEqual(w.String(), "123456").NotEmpty(w.String())

	zstdR, err := zstd.NewReader(w)
	a.NotError(err).NotNil(zstdR)
	data, err = io.ReadAll(zstdR)
	a.NotError(err).NotNil(data).Equal(string(data), "123456")
}

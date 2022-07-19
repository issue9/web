// SPDX-License-Identifier: MIT

package encoding

import (
	"bytes"
	"compress/gzip"
	"compress/lzw"
	"io"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/issue9/assert/v2"
)

func TestEncodings_Add(t *testing.T) {
	a := assert.New(t, false)

	e := NewEncodings(nil)
	a.NotNil(e)

	e.Add("gzip-1", "gzip", GZipWriter(gzip.DefaultCompression))
	e.Add("gzip-2", "gzip", GZipWriter(gzip.DefaultCompression))
	a.Equal(2, len(e.pools))

	// 重复添加
	a.PanicString(func() {
		e.Add("gzip-1", "gzip", GZipWriter(gzip.DefaultCompression))
	}, "存在相同 ID gzip-1 的函数")

	a.PanicString(func() {
		e.Add("gzip-3", "gzip", nil)
	}, "参数 w 不能为空")

	a.PanicString(func() {
		e.Add("gzip-3", "*", GZipWriter(gzip.DefaultCompression))
	}, "name 值不能为 identity 和 *")

	a.PanicString(func() {
		e.Add("gzip-3", "identity", GZipWriter(gzip.DefaultCompression))
	}, "name 值不能为 identity 和 *")
}

func TestEncodings_Allow(t *testing.T) {
	a := assert.New(t, false)
	e := NewEncodings(nil)
	a.NotNil(e)

	e.Add("lzw-lsb-2", "compress", CompressWriter(lzw.LSB, 2))
	e.Add("lzw-msb-2", "compress", CompressWriter(lzw.MSB, 2))
	e.Add("lzw-msb-5", "compress", CompressWriter(lzw.MSB, 5))
	e.Add("gzip-3", "gzip", GZipWriter(3))
	e.Add("gzip-9", "gzip", GZipWriter(9))
	e.Add("deflate-9", "deflate", DeflateWriter(9))

	a.PanicString(func() {
		e.Allow("text/html")
	}, "id 不能为空")

	a.PanicString(func() {
		e.Allow("text/html", "not-exists")
	}, "未找到 id 为 not-exists 表示的算法")

	a.PanicString(func() {
		e.Allow("text/html", "lzw-lsb-2", "lzw-msb-2") // 都是 compress
	}, "id 引用中存在多个名为 compress 的算法")

	a.Empty(e.allowTypes).
		Empty(e.allowTypesPrefix)

	e.Allow("text/html", "lzw-lsb-2", "gzip-3")
	a.Length(e.allowTypes, 1).
		Length(e.allowTypes["text/html"], 2)

	a.PanicString(func() {
		e.Allow("text/html", "deflate-9")
	}, "已经存在对 text/html 的压缩规则")

	e.Allow("application/*", "lzw-lsb-2", "gzip-3")
	a.Length(e.allowTypes, 1).
		Length(e.allowTypesPrefix, 1).
		Length(e.allowTypesPrefix[0].pools, 2)

	a.PanicString(func() {
		e.Allow("application/*", "deflate-9")
	}, "已经存在对 application/* 的压缩规则")
}

func TestEncodings_Search(t *testing.T) {
	a := assert.New(t, false)

	e := NewEncodings(nil)
	a.NotNil(e)

	e.Add("lzw-lsb-2", "compress", CompressWriter(lzw.LSB, 2))
	e.Add("lzw-msb-2", "compress", CompressWriter(lzw.MSB, 2))
	e.Add("lzw-msb-5", "compress", CompressWriter(lzw.MSB, 5))
	e.Add("gzip-3", "gzip", GZipWriter(3))
	e.Add("gzip-9", "gzip", GZipWriter(9))
	e.Add("deflate-9", "deflate", DeflateWriter(9))

	e.Allow("text/plain", "lzw-lsb-2", "gzip-3")
	e.Allow("application/*", "gzip-9", "lzw-lsb-2")

	a.Run("一般", func(a *assert.Assertion) {
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

	a.Run("header=*", func(a *assert.Assertion) {
		b, notAccept := e.Search("application/xml", "*;q=0")
		a.True(notAccept).Nil(b)

		b, notAccept = e.Search("application/xml", "*,br")
		a.False(notAccept).NotNil(b).Equal(b.name, "gzip")

		b, notAccept = e.Search("application/xml", "*,gzip")
		a.False(notAccept).NotNil(b).Equal(b.name, "compress")

		b, notAccept = e.Search("application/xml", "*,gzip,compress") // gzip,compress 都排除了
		a.False(notAccept).Nil(b)
	})

	a.Run("header=identity", func(a *assert.Assertion) {
		b, notAccept := e.Search("application/xml", "identity,gzip,br")
		a.False(notAccept).NotNil(b).Equal(b.name, "gzip")

		// 正常匹配
		b, notAccept = e.Search("application/xml", "identity;q=0,gzip,br")
		a.False(notAccept).NotNil(b).Equal(b.name, "gzip")

		// 没有可匹配，选取第一个
		b, notAccept = e.Search("application/xml", "identity;q=0,abc,def")
		a.False(notAccept).Nil(b)
	})

	a.Run("同时匹配多条 contentType", func(a *assert.Assertion) {
		b, notAccept := e.Search("image/jpg", "gzip") // image/jpg 不匹配任何 content-type
		a.False(notAccept).Nil(b)

		e.Allow("*", "deflate-9")
		e.Allow("application*", "deflate-9")

		b, notAccept = e.Search("image/jpg", "deflate") // 匹配 *
		a.False(notAccept).NotNil(b).Equal(b.name, "deflate")

		b, notAccept = e.Search("application/xml", "identity,gzip,br") // application/xml 依然遵照 application/* 匹配
		a.False(notAccept).NotNil(b).Equal(b.name, "gzip")

		b, notAccept = e.Search("application1", "identity,gzip,br") // 匹配 application*
		a.False(notAccept).NotNil(b).Equal(b.name, "deflate")
	})
}

func TestEncodings_Compress(t *testing.T) {
	a := assert.New(t, false)

	e := NewEncodings(nil)
	a.NotNil(e)

	e.Add("lzw-lsb-2", "compress", CompressWriter(lzw.LSB, 2))
	e.Add("lzw-msb-2", "compress", CompressWriter(lzw.MSB, 2))
	e.Add("lzw-msb-5", "compress", CompressWriter(lzw.MSB, 5))
	e.Add("gzip-3", "gzip", GZipWriter(3))
	e.Add("gzip-9", "gzip", GZipWriter(9))
	e.Add("br-3-10", "br", BrotliWriter(brotli.WriterOptions{Quality: 3, LGWin: 10}))

	e.Allow("text/*", "lzw-lsb-2", "gzip-3")
	e.Allow("application/*", "lzw-lsb-2", "gzip-3", "br-3-10")

	b, notAccept := e.Search("application/json", "gzip;q=0.9,br")
	a.False(notAccept).NotNil(b).Equal(b.name, "br")

	w := &bytes.Buffer{}
	wc := b.Get(w)
	_, err := wc.Write([]byte("123456"))
	a.NotError(err)
	a.NotError(wc.Close())
	a.NotEqual(w.String(), "123456").NotEmpty(w.String())

	r := brotli.NewReader(w)
	a.NotNil(r)
	data, err := io.ReadAll(r)
	a.NotError(err).NotNil(data).Equal(string(data), "123456")
}

// SPDX-License-Identifier: MIT

package codec

import (
	"bytes"
	"compress/gzip"
	"compress/lzw"
	"io"
	"testing"

	"github.com/issue9/assert/v3"
)

func TestCodec_ContentEncoding(t *testing.T) {
	a := assert.New(t, false)
	e := New()
	a.NotNil(e)

	e.AddCompressor("compress", NewLZWCompressor(lzw.LSB, 2), "text/plain", "application/*").
		AddCompressor("gzip", NewGzipCompressor(3), "text/plain").
		AddCompressor("gzip", NewGzipCompressor(9), "application/*")

	r := &bytes.Buffer{}
	gw := gzip.NewWriter(r)
	_, err := gw.Write([]byte("123"))
	a.NotError(err).
		NotError(gw.Flush()).
		NotError(gw.Close())

	rr, err := e.ContentEncoding("gzip", r)
	a.NotError(err).NotNil(rr)
	data, err := io.ReadAll(rr)
	a.NotError(err).Equal(string(data), "123")
}

func TestCodec_AcceptEncoding(t *testing.T) {
	a := assert.New(t, false)

	e := New()
	e.AddCompressor("compress", NewLZWCompressor(lzw.LSB, 2), "text/plain", "application/*").
		AddCompressor("gzip", NewGzipCompressor(3), "text/plain").
		AddCompressor("gzip", NewGzipCompressor(9), "application/*")

	a.Equal(e.AcceptEncodingHeader(), "compress,gzip")

	t.Run("一般", func(t *testing.T) {
		a := assert.New(t, false)
		b, name, notAccept := e.AcceptEncoding("application/json", "gzip;q=0.9,br", nil)
		a.False(notAccept).NotNil(b).Equal(name, "gzip")

		b, name, notAccept = e.AcceptEncoding("application/json", "br,gzip", nil)
		a.False(notAccept).NotNil(b).Equal(name, "gzip")

		b, name, notAccept = e.AcceptEncoding("text/plain", "gzip,br", nil)
		a.False(notAccept).NotNil(b).Equal(name, "gzip")

		b, _, notAccept = e.AcceptEncoding("text/plain", "br", nil)
		a.False(notAccept).Nil(b)

		b, _, notAccept = e.AcceptEncoding("text/plain", "", nil)
		a.False(notAccept).Nil(b)
	})

	t.Run("header=*", func(t *testing.T) {
		a := assert.New(t, false)
		b, _, notAccept := e.AcceptEncoding("application/xml", "*;q=0", nil)
		a.True(notAccept).Nil(b)

		b, name, notAccept := e.AcceptEncoding("application/xml", "*,br", nil)
		a.False(notAccept).NotNil(b).Equal(name, "compress")

		b, name, notAccept = e.AcceptEncoding("application/xml", "*,gzip", nil)
		a.False(notAccept).NotNil(b).Equal(name, "compress")

		b, _, notAccept = e.AcceptEncoding("application/xml", "*,gzip,compress", nil) // gzip,compress 都排除了
		a.False(notAccept).Nil(b)
	})

	t.Run("header=identity", func(t *testing.T) {
		a := assert.New(t, false)
		b, name, notAccept := e.AcceptEncoding("application/xml", "identity,gzip,br", nil)
		a.False(notAccept).NotNil(b).Equal(name, "gzip")

		// 正常匹配
		b, name, notAccept = e.AcceptEncoding("application/xml", "identity;q=0,gzip,br", nil)
		a.False(notAccept).NotNil(b).Equal(name, "gzip")

		// 没有可匹配，选取第一个
		b, _, notAccept = e.AcceptEncoding("application/xml", "identity;q=0,abc,def", nil)
		a.False(notAccept).Nil(b)
	})
}

// SPDX-License-Identifier: MIT

package codec

import (
	"compress/lzw"
	"testing"

	"github.com/issue9/assert/v3"

	"github.com/issue9/web/codec/compressor"
)

func TestCodec_AcceptEncoding(t *testing.T) {
	a := assert.New(t, false)

	e := New(APIMimetypes(), []*Compression{
		{Name: "compress", Compressor: compressor.NewLZWCompressor(lzw.LSB, 2), Types: []string{"text/plain", "application/*"}},
		{Name: "gzip", Compressor: compressor.NewGzipCompressor(3), Types: []string{"text/plain"}},
		{Name: "gzip", Compressor: compressor.NewGzipCompressor(9), Types: []string{"application/*"}},
	})

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

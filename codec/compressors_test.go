// SPDX-License-Identifier: MIT

package codec

import (
	"bytes"
	"compress/lzw"
	"io"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/issue9/assert/v3"
)

func TestBrotli(t *testing.T) {
	a := assert.New(t, false)
	testCompress(a, NewBrotliCompressor(brotli.WriterOptions{Quality: 6}), NewBrotliCompressor(brotli.WriterOptions{Quality: 11}))
}

func TestGzip(t *testing.T) {
	a := assert.New(t, false)
	testCompress(a, NewGzipCompressor(3), NewGzipCompressor(5))
}

func TestDeflate(t *testing.T) {
	a := assert.New(t, false)
	testCompress(a, NewDeflateCompressor(3, nil), NewDeflateCompressor(5, nil))
}

func TestLZW(t *testing.T) {
	a := assert.New(t, false)
	// lzw 的 reader 不通用。
	testCompress(a, NewLZWCompressor(lzw.LSB, 8), nil)
}

func TestZstd(t *testing.T) {
	a := assert.New(t, false)
	testCompress(a, NewZstdCompressor(), NewZstdCompressor())
}

// c1 与 c2 应该是由不同的参数初始化的，会测试相互能读取。
// 如果 c2 为 nil，表示不测试 reader 的通用性。
func testCompress(a *assert.Assertion, c1, c2 Compressor) {
	a.NotNil(c1)

	// c1 encode

	b1 := &bytes.Buffer{}
	w, err := c1.NewEncoder(b1)
	a.NotError(err).NotNil(w)
	_, err = w.Write([]byte("123"))
	a.NotError(err)
	if f, ok := w.(interface{ Flush() error }); ok {
		a.NotError(f.Flush())
	}
	a.NotError(w.Close())
	a.True(b1.Len() > 0)

	b2 := &bytes.Buffer{}
	_, err = b2.Write(b1.Bytes())
	a.NotError(err)

	// c1 read c1.encode

	r, err := c1.NewDecoder(b1)
	a.NotError(err).NotNil(r)
	data, err := io.ReadAll(r)
	a.NotError(err).Equal(string(data), "123")
	a.NotError(r.Close())

	// c2 read c1.encode

	if c2 == nil {
		return
	}

	r, err = c2.NewDecoder(b2)
	a.NotError(err).NotNil(r)
	data, err = io.ReadAll(r)
	a.NotError(err).Equal(string(data), "123")
	a.NotError(r.Close())
}

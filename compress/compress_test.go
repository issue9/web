// SPDX-License-Identifier: MIT

package compress

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
	testCompress(a, NewBrotliCompress(brotli.WriterOptions{}))
}

func TestGzip(t *testing.T) {
	a := assert.New(t, false)
	testCompress(a, NewGzipCompress(3))
}

func TestDeflate(t *testing.T) {
	a := assert.New(t, false)
	testCompress(a, NewDeflateCompress(3, nil))
}

func TestLZW(t *testing.T) {
	a := assert.New(t, false)
	testCompress(a, NewLZWCompress(lzw.LSB, 8))
}

func TestZstd(t *testing.T) {
	a := assert.New(t, false)
	testCompress(a, NewZstdCompress())
}

func testCompress(a *assert.Assertion, c Compress) {
	buf := &bytes.Buffer{}
	a.NotNil(c)

	w, err := c.Encoder(buf)
	a.NotError(err).NotNil(w)
	_, err = w.Write([]byte("123"))
	a.NotError(err)
	if f, ok := w.(interface{ Flush() error }); ok {
		a.NotError(f.Flush())
	}
	a.NotError(w.Close())
	a.True(buf.Len() > 0)

	r, err := c.Decoder(buf)
	a.NotError(err).NotNil(r)
	data, err := io.ReadAll(r)
	a.NotError(err).Equal(string(data), "123")
	a.NotError(r.Close())
}

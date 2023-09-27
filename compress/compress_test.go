// SPDX-License-Identifier: MIT

package compress

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

func TestCompress(t *testing.T) {
	a := assert.New(t, false)
	e := NewCompresses(4)
	e.Add("compress", NewLZWCompress(lzw.LSB, 2), "text/*", "application/*", "application/xml").
		Add("gzip", NewGzipCompress(3), "text/*", "application/*").
		Add("br", NewBrotliCompress(brotli.WriterOptions{Quality: 3, LGWin: 10}), "application/*").
		Add("zstd", NewZstdCompress(nil, nil), "application/xml")

	// br

	b, notAccept := e.AcceptEncoding("application/json", "gzip;q=0.9,br", nil)
	a.False(notAccept).NotNil(b).Equal(b.name, "br")

	w := &bytes.Buffer{}
	wc, err := b.Compress().Encoder(w)
	a.NotError(err).NotNil(wc)
	_, err = wc.Write([]byte("123456"))
	a.NotError(err)
	a.NotError(wc.Close())
	a.NotEqual(w.String(), "123456").NotEmpty(w.String())

	brotliR := brotli.NewReader(w)
	a.NotNil(brotliR)
	data, err := io.ReadAll(brotliR)
	a.NotError(err).NotNil(data).Equal(string(data), "123456")

	// gzip

	b, notAccept = e.AcceptEncoding("application/json", "gzip;q=0.9", nil)
	a.False(notAccept).NotNil(b).Equal(b.name, "gzip")

	w = &bytes.Buffer{}
	wc, err = b.Compress().Encoder(w)
	a.NotError(err).NotNil(wc)
	_, err = wc.Write([]byte("123456"))
	a.NotError(err)
	a.NotError(wc.Close())
	a.NotEqual(w.String(), "123456").NotEmpty(w.String())

	gzipR, err := gzip.NewReader(w)
	a.NotError(err).NotNil(gzipR)
	data, err = io.ReadAll(gzipR)
	a.NotError(err).NotNil(data).Equal(string(data), "123456")

	// zstd

	b, notAccept = e.AcceptEncoding("application/xml", "zstd", nil)
	a.False(notAccept).NotNil(b).Equal(b.name, "zstd")

	w = &bytes.Buffer{}
	wc, err = b.Compress().Encoder(w)
	a.NotError(err).NotNil(wc)
	_, err = wc.Write([]byte("123456"))
	a.NotError(err)
	a.NotError(wc.Close())
	a.NotEqual(w.String(), "123456").NotEmpty(w.String())

	zstdR, err := zstd.NewReader(w)
	a.NotError(err).NotNil(zstdR)
	data, err = io.ReadAll(zstdR)
	a.NotError(err).NotNil(data).Equal(string(data), "123456")
}

// SPDX-License-Identifier: MIT

package web

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

func TestServer_searchAlg(t *testing.T) {
	a := assert.New(t, false)
	srv := newTestServer(a, &Options{
		Encodings: []*Encoding{
			{Name: "compress", Builder: CompressWriter(lzw.LSB, 2), ContentTypes: []string{"text/plain", "application/*"}},
			{Name: "gzip", Builder: GZipWriter(3), ContentTypes: []string{"text/plain"}},
			{Name: "gzip", Builder: GZipWriter(9), ContentTypes: []string{"application/*"}},
		},
	})

	t.Run("一般", func(t *testing.T) {
		a := assert.New(t, false)
		b, notAccept := srv.searchAlg("application/json", "gzip;q=0.9,br")
		a.False(notAccept).NotNil(b).Equal(b.name, "gzip")

		b, notAccept = srv.searchAlg("application/json", "br,gzip")
		a.False(notAccept).NotNil(b).Equal(b.name, "gzip")

		b, notAccept = srv.searchAlg("text/plain", "gzip,br")
		a.False(notAccept).NotNil(b).Equal(b.name, "gzip")

		b, notAccept = srv.searchAlg("text/plain", "br")
		a.False(notAccept).Nil(b)

		b, notAccept = srv.searchAlg("text/plain", "")
		a.False(notAccept).Nil(b)
	})

	t.Run("header=*", func(t *testing.T) {
		a := assert.New(t, false)
		b, notAccept := srv.searchAlg("application/xml", "*;q=0")
		a.True(notAccept).Nil(b)

		b, notAccept = srv.searchAlg("application/xml", "*,br")
		a.False(notAccept).NotNil(b).Equal(b.name, "compress")

		b, notAccept = srv.searchAlg("application/xml", "*,gzip")
		a.False(notAccept).NotNil(b).Equal(b.name, "compress")

		b, notAccept = srv.searchAlg("application/xml", "*,gzip,compress") // gzip,compress 都排除了
		a.False(notAccept).Nil(b)
	})

	t.Run("header=identity", func(t *testing.T) {
		a := assert.New(t, false)
		b, notAccept := srv.searchAlg("application/xml", "identity,gzip,br")
		a.False(notAccept).NotNil(b).Equal(b.name, "gzip")

		// 正常匹配
		b, notAccept = srv.searchAlg("application/xml", "identity;q=0,gzip,br")
		a.False(notAccept).NotNil(b).Equal(b.name, "gzip")

		// 没有可匹配，选取第一个
		b, notAccept = srv.searchAlg("application/xml", "identity;q=0,abc,def")
		a.False(notAccept).Nil(b)
	})
}

func TestCompress(t *testing.T) {
	a := assert.New(t, false)
	srv := newTestServer(a, &Options{
		Encodings: []*Encoding{
			{Name: "compress", Builder: CompressWriter(lzw.LSB, 2), ContentTypes: []string{"text/*", "application/*", "application/xml"}},
			{Name: "gzip", Builder: GZipWriter(3), ContentTypes: []string{"text/*", "application/*"}},
			{Name: "br", Builder: BrotliWriter(brotli.WriterOptions{Quality: 3, LGWin: 10}), ContentTypes: []string{"application/*"}},
			{Name: "zstd", Builder: ZstdWriter(zstd.WithEncoderLevel(zstd.SpeedDefault)), ContentTypes: []string{"application/xml"}},
		},
	})

	// br

	b, notAccept := srv.searchAlg("application/json", "gzip;q=0.9,br")
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

	b, notAccept = srv.searchAlg("application/json", "gzip;q=0.9")
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

	b, notAccept = srv.searchAlg("application/xml", "zstd")
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

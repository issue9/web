// SPDX-License-Identifier: MIT

package compress

import (
	"bytes"
	"compress/gzip"
	"testing"

	"github.com/issue9/assert/v3"
)

func BenchmarkCompress_Encoder(b *testing.B) {
	a := assert.New(b, false)
	e := NewCompresses(3).Add("gzip", NewGzipCompress(3), "application/*")

	pool, notAccept := e.AcceptEncoding("application/json", "gzip,br;q=0.9", nil)
	a.False(notAccept).NotNil(pool).Equal(pool.name, "gzip")
	w := &bytes.Buffer{}

	for i := 0; i < b.N; i++ {
		w.Reset()

		wc, err := pool.Compress().Encoder(w)
		a.NotError(err).
			NotNil(wc).
			NotError(wc.Close())
	}
}

func BenchmarkCompress_Decoder(b *testing.B) {
	a := assert.New(b, false)
	e := NewCompresses(3).Add("gzip", NewGzipCompress(3), "application/*")

	pool, notAccept := e.AcceptEncoding("application/json", "gzip,br;q=0.9", nil)
	a.False(notAccept).NotNil(pool).Equal(pool.name, "gzip")

	r := &bytes.Buffer{}
	gw := gzip.NewWriter(r)
	_, err := gw.Write([]byte(""))
	a.NotError(err).
		NotError(gw.Flush()).
		NotError(gw.Close())
	data := r.Bytes()

	for i := 0; i < b.N; i++ {
		wc, err := pool.Compress().Decoder(bytes.NewBuffer(data))
		a.NotError(err).
			NotNil(wc).
			NotError(wc.Close())
	}
}

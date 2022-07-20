// SPDX-License-Identifier: MIT

package encoding

import (
	"bytes"
	"compress/gzip"
	"io"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/issue9/assert/v2"
)

func BenchmarkPool_Get(b *testing.B) {
	a := assert.New(b, false)

	e := NewEncodings(nil)
	a.NotNil(e)
	e.Add("gzip-3", "gzip", GZipWriter(3))
	e.Add("gzip-9", "gzip", GZipWriter(9))
	e.Add("br-3-10", "br", BrotliWriter(brotli.WriterOptions{Quality: 3, LGWin: 10}))
	e.Allow("application/*", "gzip-3", "br-3-10")

	pool, notAccept := e.Search("application/json", "gzip,br;q=0.9")
	a.False(notAccept).NotNil(pool).Equal(pool.name, "gzip")

	for i := 0; i < b.N; i++ {
		w := &bytes.Buffer{}
		wc := pool.Get(w)
		_, err := wc.Write([]byte("123456"))
		a.NotError(err)
		a.NotError(wc.Close())

		r, err := gzip.NewReader(w)
		a.NotError(err).NotNil(r)
		data, err := io.ReadAll(r)
		a.NotError(err).NotNil(data).Equal(string(data), "123456")
	}
}

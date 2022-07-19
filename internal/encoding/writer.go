// SPDX-License-Identifier: MIT

package encoding

import (
	"compress/flate"
	"compress/gzip"
	"compress/lzw"
	"io"

	"github.com/andybalholm/brotli"
)

//  WriteCloseRester 每种压缩实例需要实现的最小接口
type WriteCloseRester interface {
	io.WriteCloser
	Reset(io.Writer)
}

type NewEncodingFunc func() WriteCloseRester

type compressWriter struct {
	*lzw.Writer
	order lzw.Order
	width int
}

func newCompressWriter(w io.Writer, order lzw.Order, width int) *compressWriter {
	return &compressWriter{
		Writer: lzw.NewWriter(w, order, width).(*lzw.Writer),
	}
}

func (cw *compressWriter) Reset(w io.Writer) {
	cw.Writer.Reset(w, cw.order, cw.width)
}

// GZipWriter gzip
func GZipWriter() WriteCloseRester {
	return gzip.NewWriter(nil)
}

// DeflateWriter deflate
func DeflateWriter() WriteCloseRester {
	w, err := flate.NewWriter(nil, flate.DefaultCompression)
	if err != nil {
		panic(err)
	}
	return w
}

// BrotliWriter br
func BrotliWriter() WriteCloseRester {
	return brotli.NewWriter(nil)
}

// CompressWriter compress
func CompressWriter() WriteCloseRester {
	return newCompressWriter(nil, lzw.LSB, 5)
}

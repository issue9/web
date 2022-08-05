// SPDX-License-Identifier: MIT

package encoding

import (
	"compress/flate"
	"compress/gzip"
	"compress/lzw"
	"io"

	"github.com/andybalholm/brotli"
)

// WriteCloseRester 每种压缩实例需要实现的最小接口
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

func (cw *compressWriter) Reset(w io.Writer) {
	cw.Writer.Reset(w, cw.order, cw.width)
}

// GZipWriter gzip
func GZipWriter(level int) NewEncodingFunc {
	return func() WriteCloseRester {
		w, err := gzip.NewWriterLevel(nil, level)
		if err != nil {
			panic(err)
		}
		return w
	}
}

// DeflateWriter deflate
func DeflateWriter(level int) NewEncodingFunc {
	return func() WriteCloseRester {
		w, err := flate.NewWriter(nil, level)
		if err != nil {
			panic(err)
		}
		return w
	}
}

// BrotliWriter br
func BrotliWriter(o brotli.WriterOptions) NewEncodingFunc {
	return func() WriteCloseRester {
		return brotli.NewWriterOptions(nil, o)
	}
}

// CompressWriter compress
func CompressWriter(order lzw.Order, width int) NewEncodingFunc {
	return func() WriteCloseRester {
		return &compressWriter{
			Writer: lzw.NewWriter(nil, order, width).(*lzw.Writer),
		}
	}
}

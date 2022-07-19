// SPDX-License-Identifier: MIT

package encoding

import (
	"compress/flate"
	"compress/gzip"
	"compress/lzw"
	"io"

	"github.com/andybalholm/brotli"
)

//  每种压缩实例需要实现的最小接口
type WriteCloseRester interface {
	io.WriteCloser
	Reset(io.Writer)
}

type NewEncodingFunc func() (WriteCloseRester, error)

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
func GZipWriter() (WriteCloseRester, error) {
	return gzip.NewWriter(nil), nil
}

// DeflateWriter deflate
func DeflateWriter() (WriteCloseRester, error) {
	return flate.NewWriter(nil, flate.DefaultCompression)
}

// BrotliWriter br
func BrotliWriter() (WriteCloseRester, error) {
	return brotli.NewWriter(nil), nil
}

// CompressWriter compress
func CompressWriter() (WriteCloseRester, error) {
	return newCompressWriter(nil, lzw.LSB, 5), nil
}

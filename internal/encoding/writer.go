// SPDX-License-Identifier: MIT

package encoding

import (
	"bytes"
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
func GZipWriter(w io.Writer) (WriteCloseRester, error) {
	return gzip.NewWriter(w), nil
}

// DeflateWriter deflate
func DeflateWriter(w io.Writer) (WriteCloseRester, error) {
	return flate.NewWriter(&bytes.Buffer{}, flate.DefaultCompression)
}

// BrotliWriter br
func BrotliWriter(w io.Writer) (WriteCloseRester, error) {
	return brotli.NewWriter(w), nil
}

// CompressWriter compress
func CompressWriter(w io.Writer) (WriteCloseRester, error) {
	return newCompressWriter(w, lzw.LSB, 5), nil
}

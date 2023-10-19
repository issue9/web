// SPDX-License-Identifier: MIT

//go:build ignore

package main

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/lzw"
	"strconv"

	"github.com/andybalholm/brotli"
	"github.com/issue9/errwrap"
	"github.com/issue9/source"
	"github.com/klauspost/compress/zstd"

	"github.com/issue9/web/internal/problems/make"
)

const (
	pkg  = "compressor"
	file = "readers_data.go"
)

func main() {
	buf := &bytes.Buffer{}

	b := &errwrap.Buffer{}

	b.WString(make.FileHeader).
		WString("package ").WString(pkg).WString("\n\n")

	b.WString("var (\n")

	// zstd
	buf.Reset()
	zw, err := zstd.NewWriter(buf)
	checkErr(err)
	checkErr(zw.Flush())
	checkErr(zw.Close())
	b.WString("zstdInitData=[]byte {")
	if buf.Len() > 0 {
		for _, bb := range buf.Bytes() {
			b.WString(strconv.Itoa(int(bb))).WByte(',')
		}
		b.Truncate(b.Len() - 1)
	}
	b.WString("}\n")

	// gzip
	buf.Reset()
	gw := gzip.NewWriter(buf)
	checkErr(gw.Flush())
	checkErr(gw.Close())
	b.WString("gzipInitData=[]byte {")
	if buf.Len() > 0 {
		for _, bb := range buf.Bytes() {
			b.WString(strconv.Itoa(int(bb))).WByte(',')
		}
		b.Truncate(b.Len() - 1)
	}
	b.WString("}\n")

	// deflate
	buf.Reset()
	fw, err := flate.NewWriter(buf, flate.BestCompression)
	checkErr(err)
	checkErr(fw.Flush())
	checkErr(fw.Close())
	b.WString("deflateInitData=[]byte {")
	if buf.Len() > 0 {
		for _, bb := range buf.Bytes() {
			b.WString(strconv.Itoa(int(bb))).WByte(',')
		}
		b.Truncate(b.Len() - 1)
	}
	b.WString("}\n")

	// lzw
	buf.Reset()
	lw := lzw.NewWriter(buf, lzw.LSB, 2)
	checkErr(lw.Close())
	b.WString("lzwInitData=[]byte {")
	if buf.Len() > 0 {
		for _, bb := range buf.Bytes() {
			b.WString(strconv.Itoa(int(bb))).WByte(',')
		}
		b.Truncate(b.Len() - 1)
	}
	b.WString("}\n")

	// br
	buf.Reset()
	bw := brotli.NewWriter(buf)
	checkErr(bw.Flush())
	checkErr(bw.Close())
	b.WString("brotliInitData=[]byte {")
	if buf.Len() > 0 {
		for _, bb := range buf.Bytes() {
			b.WString(strconv.Itoa(int(bb))).WByte(',')
		}
		b.Truncate(b.Len() - 1)
	}
	b.WString("}\n")

	b.WString(")\n\n")

	checkErr(b.Err)
	checkErr(source.DumpGoSource(file, b.Bytes()))
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

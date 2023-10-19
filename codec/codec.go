// SPDX-License-Identifier: MIT

//go:generate go run ./make_data.go

// Package codec 编码解码工具
//
// 包含了压缩方法和媒体类型的处理。
package codec

type Codec struct {
	compresses           []*namedCompressor
	acceptEncodingHeader string
	disableCompress      bool

	types        []*mimetype
	acceptHeader string
}

func New() *Codec {
	return &Codec{
		compresses: make([]*namedCompressor, 0, 10),
		types:      make([]*mimetype, 0, 10),
	}
}

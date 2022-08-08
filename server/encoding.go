// SPDX-License-Identifier: MIT

package server

import (
	"compress/lzw"

	"github.com/andybalholm/brotli"

	"github.com/issue9/web/internal/encoding"
)

type NewEncodingFunc = encoding.NewEncodingFunc

type Encodings interface {
	// Add 添加压缩算法
	//
	// id 表示当前算法的唯一名称，在 Allow 中可以用来查找使用；
	// name 表示通过 Accept-Encoding 匹配的名称；
	// f 表示生成压缩对象的方法；
	Add(id, name string, f NewEncodingFunc)

	// Allow 允许 contentType 采用的压缩方式
	//
	// id 是指由 Add 中指定的值；
	// contentType 表示经由 Accept-Encoding 提交的值，该值不能是 identity 和 *；
	//
	// 如果未添加任何算法，则每个请求都相当于是 identity 规则。
	Allow(contentType string, id ...string)
}

// Encodings 返回压缩编码管理
func (srv *Server) Encodings() Encodings { return srv.encodings }

// EncodingGZip 返回指定配置的 gzip 算法
func EncodingGZip(level int) NewEncodingFunc { return encoding.GZipWriter(level) }

// EncodingDeflate 返回指定配置的 deflate 算法
func EncodingDeflate(level int) NewEncodingFunc { return encoding.DeflateWriter(level) }

// EncodingBrotli 返回指定配置的 br 算法
func EncodingBrotli(o brotli.WriterOptions) NewEncodingFunc { return encoding.BrotliWriter(o) }

// EncodingCompress 返回指定配置的 compress 算法
func EncodingCompress(order lzw.Order, width int) NewEncodingFunc {
	return encoding.CompressWriter(order, width)
}

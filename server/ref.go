// SPDX-License-Identifier: MIT

package server

import (
	"compress/lzw"

	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/zstd"

	"github.com/issue9/web/internal/encoding"
	"github.com/issue9/web/internal/files"
	"github.com/issue9/web/internal/service"
)

type (
	Services        = service.Server
	NewEncodingFunc = encoding.NewEncodingFunc
	Files           = files.Files

	// MarshalFunc 序列化函数原型
	MarshalFunc func(*Context, any) ([]byte, error)

	// UnmarshalFunc 反序列化函数原型
	UnmarshalFunc func([]byte, any) error

	Encodings interface {
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

	Mimetypes interface {
		// Exists 是否存在同名的
		Exists(string) bool

		// Delete 删除指定名称的编码方法
		Delete(string)

		// Add 添加新的编码方法
		//
		// name 为编码名称；
		// problem 为该编码在返回 [Problem] 对象时的 mimetype 报头值，如果为空，则会被赋予 name 相同的值；
		Add(name string, m MarshalFunc, u UnmarshalFunc, problem string)

		// Set 修改指定名称的相关配置
		//
		// name 用于查找相关的编码方法；
		// 如果 problem 为空，会被赋予与 name 相同的值；
		Set(name string, m MarshalFunc, u UnmarshalFunc, problem string)

		Len() int
	}
)

// Files 配置文件的相关操作
func (srv *Server) Files() *Files { return srv.files }

// Services 服务管理
func (srv *Server) Services() *Services { return srv.services }

// Mimetypes 编解码控制
func (srv *Server) Mimetypes() Mimetypes { return srv.mimetypes }

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

// EncodingZstd 返回指定配置的 zstd 算法
func EncodingZstd(o ...zstd.EOption) NewEncodingFunc { return encoding.ZstdWriter(o...) }

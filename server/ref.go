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
	//
	// NOTE: MarshalFunc 的实现中不能调用 [Context.Marshal] 方法。
	MarshalFunc func(*Context, any) ([]byte, error)

	// UnmarshalFunc 反序列化函数原型
	UnmarshalFunc func([]byte, any) error

	Mimetypes interface {
		// Exists 是否存在同名的
		Exists(string) bool

		// Delete 删除指定名称的编码方法
		Delete(string)

		// Add 添加新的编码方法
		//
		// name 为编码名称；
		// problem 为该编码在返回 [Problem] 对象时的 mimetype 报头值，
		// 如果为空，则会被赋予 name 相同的值；
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
//
// 在 [Server] 初始之后，所有的服务就处于运行状态，
// 后续添加的服务也会自动运行。
func (srv *Server) Services() *Services { return srv.services }

// Mimetypes 编解码控制
func (srv *Server) Mimetypes() Mimetypes { return srv.mimetypes }

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

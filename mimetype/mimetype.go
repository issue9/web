// SPDX-License-Identifier: MIT

// Package mimetype 提供了对编码的支持。
package mimetype

// DefaultMimetype 默认的媒体类型，在不能获取输入和输出的媒体类型时，
// 会采用此值作为其默认值。
//
// 若编码函数中指定该类型的函数，则会使用该编码优先匹配 */* 等格式的请求。
const DefaultMimetype = "application/octet-stream"

// Nil 表示向客户端输出 nil 值。
//
// 这是一个只有类型但是值为空的变量。在某些特殊情况下，
// 如果需要向客户端输出一个 nil 值的内容，可以使用此值。
var Nil *struct{}

// MarshalFunc 将一个对象转换成 []byte 内容时，所采用的接口。
type MarshalFunc func(v interface{}) ([]byte, error)

// UnmarshalFunc 将客户端内容转换成一个对象时，所采用的接口。
type UnmarshalFunc func([]byte, interface{}) error

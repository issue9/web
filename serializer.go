// SPDX-License-Identifier: MIT

package web

import "github.com/issue9/web/internal/mimetypes"

var errUnsupportedSerialization = NewLocaleError("unsupported serialization")

type (
	// BuildMarshalFunc 生成特定于 [Context] 的 [MarshalFunc]
	//
	// 如果传递的参数是空值，应该返回一个默认的 [MarshalFunc] 实现，
	// 该实现将被用于 [Client] 的相关功能。
	BuildMarshalFunc func(*Context) MarshalFunc // 不能是 alias，// https://github.com/golang/go/issues/50729

	// MarshalFunc 序列化函数原型
	//
	// NOTE: MarshalFunc 的作用是输出内容，所以在实现中不能调用 [Context.Render] 等输出方法。
	MarshalFunc = func(any) ([]byte, error)

	// UnmarshalFunc 反序列化函数原型
	UnmarshalFunc = func([]byte, any) error

	mtsType = mimetypes.Mimetypes[BuildMarshalFunc, UnmarshalFunc]

	mtType = mimetypes.Mimetype[BuildMarshalFunc, UnmarshalFunc]
)

// ErrUnsupportedSerialization 返回不支持序列化的错误信息
func ErrUnsupportedSerialization() error { return errUnsupportedSerialization }

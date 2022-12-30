// SPDX-License-Identifier: MIT

// Package serializer 序列化相关的接口定义
package serializer

import (
	"io/fs"

	"github.com/issue9/localeutil"
)

// ErrUnsupported 返回不支持序列化的错误信息
//
// 当一个对象无法被正常的序列化或是反序列化时返回此错误。
var ErrUnsupported = localeutil.Error("unsupported serialization")

type (
	// Serializer 管理注册的序列化函数
	Serializer interface {
		// Items 返回所有的注册项名称
		Items() []string

		// Exists 是否指定名称的项
		Exists(string) bool

		// Add 添加序列化函数
		//
		// m 和 u 可以为 nil，表示仅作为一个占位符使用；
		//
		// name 表示之后用于查找该序列化函数的唯一 ID，
		// 后期用户可以根据 name 从 Search 直接查找相应的序列化函数。
		Add(m MarshalFunc, u UnmarshalFunc, name ...string) error

		// Set 修改或是添加
		Set(name string, m MarshalFunc, u UnmarshalFunc)

		// Delete 删除指定名称的序列化方函数
		Delete(name string)

		// Search 根据注册时的名称查找
		Search(name string) (string, MarshalFunc, UnmarshalFunc)

		// SearchFunc 按指定方法查找序列化方法
		//
		// 如果返回的 name 为空，表示没有找到
		SearchFunc(match func(string) bool) (string, MarshalFunc, UnmarshalFunc)

		// Len 返回注册的数量
		Len() int
	}

	// MarshalFunc 序列化函数原型
	MarshalFunc func(any) ([]byte, error)

	// UnmarshalFunc 反序列化函数原型
	UnmarshalFunc func([]byte, any) error

	// FS 对文件系统的序列化支持
	FS interface {
		// Serializer 序列化函数管理接口
		Serializer() Serializer

		// Save 将 v 序列化并保存至 p
		//
		// 根据 p 后缀名选择序列化方法。
		Save(p string, v any) error

		// Load 加载文件到 v
		//
		// 根据 name 后缀名选择序列化方法。
		Load(fsys fs.FS, name string, v any) error
	}
)

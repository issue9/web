// SPDX-License-Identifier: MIT

// Package serializer 序列化相关的操作
package serializer

import (
	"io/fs"

	"github.com/issue9/localeutil"
	"github.com/issue9/sliceutil"
)

// ErrUnsupported 返回不支持序列化的错误信息
//
// 当一个对象无法被正常的序列化或是反序列化是，返回此错误。
var ErrUnsupported = localeutil.Error("unsupported serialization")

type (
	// Serializer 管理注册的序列化函数
	Serializer struct {
		serializes []*serializer
	}

	// MarshalFunc 序列化函数原型
	MarshalFunc func(v any) ([]byte, error)

	// UnmarshalFunc 反序列化函数原型
	UnmarshalFunc func([]byte, any) error

	serializer struct {
		Name      string
		Marshal   MarshalFunc
		Unmarshal UnmarshalFunc
	}

	// FS 对文件系统的序列化支持
	FS interface {
		// Serializer 序列化函数管理接口
		Serializer() *Serializer

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

func New(c int) *Serializer {
	return &Serializer{serializes: make([]*serializer, 0, c)}
}

// Add 添加序列化函数
//
// m 和 u 可以为 nil，表示仅作为一个占位符使用；
//
// name 表示之后用于查找该序列化函数的唯一 ID，
// 后期用户可以根据 name 从 c.Search 直接查找相应的序列化函数。
func (s *Serializer) Add(m MarshalFunc, u UnmarshalFunc, name ...string) error {
	for _, n := range name {
		if err := s.add(n, m, u); err != nil {
			return err
		}
	}
	return nil
}

func (s *Serializer) add(name string, m MarshalFunc, u UnmarshalFunc) error {
	for _, mt := range s.serializes {
		if mt.Name == name {
			return localeutil.Error("has serialization function %s", name)
		}
	}

	s.serializes = append(s.serializes, &serializer{
		Name:      name,
		Marshal:   m,
		Unmarshal: u,
	})

	return nil
}

// Set 修改或是添加
func (s *Serializer) Set(name string, m MarshalFunc, u UnmarshalFunc) {
	for _, mt := range s.serializes {
		if mt.Name == name {
			mt.Marshal = m
			mt.Unmarshal = u
			return
		}
	}

	s.serializes = append(s.serializes, &serializer{
		Name:      name,
		Marshal:   m,
		Unmarshal: u,
	})
}

// Delete 删除指定名称的数据
func (s *Serializer) Delete(name string) {
	s.serializes = sliceutil.Delete(s.serializes, func(e *serializer) bool {
		return e.Name == name
	})
}

// Search 根据注册时的名称查找
func (s *Serializer) Search(name string) (string, MarshalFunc, UnmarshalFunc) {
	return s.SearchFunc(func(n string) bool { return n == name })
}

// SearchFunc 按指定方法查找序列化方法
//
// 如果返回的 name 为空，表示没有找到
func (s *Serializer) SearchFunc(match func(string) bool) (string, MarshalFunc, UnmarshalFunc) {
	for _, mt := range s.serializes {
		if match(mt.Name) {
			return mt.Name, mt.Marshal, mt.Unmarshal
		}
	}
	return "", nil, nil
}

// Len 返回注册的数量
func (s *Serializer) Len() int { return len(s.serializes) }

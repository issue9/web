// SPDX-License-Identifier: MIT

// Package serialization 序列化相关的操作
package serialization

import (
	"github.com/issue9/localeutil"
	"github.com/issue9/sliceutil"
)

// ErrUnsupported 返回不支持序列化的错误信息
//
// 当一个对象无法被正常的序列化或是反序列化是，返回此错误。
var ErrUnsupported = localeutil.Error("unsupported serialization")

type (
	// Serialization 管理注册的序列化函数
	Serialization struct {
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
)

// New 声明 Serialization 对象
func New(c int) *Serialization {
	return &Serialization{serializes: make([]*serializer, 0, c)}
}

// Add 添加序列化函数
//
// m 和 u 可以为 nil，表示仅作为一个占位符使用，具体处理要在 ServeHTTP 中另作处理；
//
// name 表示之后用于查找该序列化函数的唯一 ID，
// 后期用户可以根据 name 从 c.Search 直接查找相应的序列化函数。
func (s *Serialization) Add(m MarshalFunc, u UnmarshalFunc, name ...string) error {
	for _, n := range name {
		if err := s.add(n, m, u); err != nil {
			return err
		}
	}
	return nil
}

func (s *Serialization) add(name string, m MarshalFunc, u UnmarshalFunc) error {
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
func (s *Serialization) Set(name string, m MarshalFunc, u UnmarshalFunc) {
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
func (s *Serialization) Delete(name string) {
	s.serializes = sliceutil.Delete(s.serializes, func(e *serializer) bool {
		return e.Name == name
	})
}

func (s *Serialization) Search(name string) (string, MarshalFunc, UnmarshalFunc) {
	return s.SearchFunc(func(n string) bool { return n == name })
}

// SearchFunc 如果返回的 name 为空，表示没有找到
func (s *Serialization) SearchFunc(match func(string) bool) (string, MarshalFunc, UnmarshalFunc) {
	for _, mt := range s.serializes {
		if match(mt.Name) {
			return mt.Name, mt.Marshal, mt.Unmarshal
		}
	}
	return "", nil, nil
}

// Len 返回注册的数量
func (s *Serialization) Len() int { return len(s.serializes) }

// SPDX-License-Identifier: MIT

// Package serialization 序列化相关的操作
package serialization

import (
	"fmt"

	"github.com/issue9/sliceutil"
)

type (
	Serialization struct {
		serializes []*serializer
	}

	// MarshalFunc 序列化函数原型
	MarshalFunc func(v interface{}) ([]byte, error)

	// UnmarshalFunc 反序列化函数原型
	UnmarshalFunc func([]byte, interface{}) error

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
// name 表示名称，一般为 mimetype 名称，比如 application/xml 等，用户也可以添加其它值，比如：
//  c.Add(json.Marshal, json.Unmarshal, "application/json", ".json")
// 后期用户可以根据文件后缀名从 c.Search 直接查找相应的序列化函数。
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
			return fmt.Errorf("已经存在相同名称 %s 的序列化函数", name)
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
func (s *Serialization) Set(name string, m MarshalFunc, u UnmarshalFunc) error {
	for _, mt := range s.serializes {
		if mt.Name == name {
			mt.Marshal = m
			mt.Unmarshal = u
			return nil
		}
	}

	s.serializes = append(s.serializes, &serializer{
		Name:      name,
		Marshal:   m,
		Unmarshal: u,
	})

	return fmt.Errorf("未找到指定名称 %s 的编解码函数", name)
}

// Delete 删除指定名称的数据
func (s *Serialization) Delete(name string) {
	size := sliceutil.Delete(s.serializes, func(i int) bool {
		return s.serializes[i].Name == name
	})
	s.serializes = s.serializes[:size]
}

func (s *Serialization) Search(name string) (string, MarshalFunc, UnmarshalFunc) {
	return s.SearchFunc(func(n string) bool { return n == name })
}

// 如果返回的 name 为空，表示没有找到
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

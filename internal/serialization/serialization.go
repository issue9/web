// SPDX-License-Identifier: MIT

// Package serialization 序列化相关的功能实现
package serialization

import (
	"github.com/issue9/localeutil"
	"github.com/issue9/sliceutil"

	"github.com/issue9/web/serializer"
)

type (
	Serialization struct {
		items []*Item
	}

	Item struct {
		Name      string
		Marshal   serializer.MarshalFunc
		Unmarshal serializer.UnmarshalFunc
	}
)

func New(c int) serializer.Serializer {
	return &Serialization{items: make([]*Item, 0, c)}
}

func (s *Serialization) Add(m serializer.MarshalFunc, u serializer.UnmarshalFunc, name ...string) error {
	for _, n := range name {
		if err := s.add(n, m, u); err != nil {
			return err
		}
	}
	return nil
}

func (s *Serialization) add(name string, m serializer.MarshalFunc, u serializer.UnmarshalFunc) error {
	for _, mt := range s.items {
		if mt.Name == name {
			return localeutil.Error("has serialization function %s", name)
		}
	}

	s.items = append(s.items, &Item{
		Name:      name,
		Marshal:   m,
		Unmarshal: u,
	})

	return nil
}

func (s *Serialization) Set(name string, m serializer.MarshalFunc, u serializer.UnmarshalFunc) {
	for _, mt := range s.items {
		if mt.Name == name {
			mt.Marshal = m
			mt.Unmarshal = u
			return
		}
	}

	s.items = append(s.items, &Item{
		Name:      name,
		Marshal:   m,
		Unmarshal: u,
	})
}

func (s *Serialization) Delete(name string) {
	s.items = sliceutil.Delete(s.items, func(e *Item) bool {
		return e.Name == name
	})
}

func (s *Serialization) Search(name string) (string, serializer.MarshalFunc, serializer.UnmarshalFunc) {
	return s.SearchFunc(func(n string) bool { return n == name })
}

func (s *Serialization) SearchFunc(match func(string) bool) (string, serializer.MarshalFunc, serializer.UnmarshalFunc) {
	for _, mt := range s.items {
		if match(mt.Name) {
			return mt.Name, mt.Marshal, mt.Unmarshal
		}
	}
	return "", nil, nil
}

func (s *Serialization) Len() int { return len(s.items) }

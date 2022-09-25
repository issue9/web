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

func New(c int) serializer.Serializer { return &Serialization{items: make([]*Item, 0, c)} }

func (s *Serialization) Items() []string {
	items := make([]string, 0, len(s.items))
	for _, item := range s.items {
		items = append(items, item.Name)
	}
	return items
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
	if s.Exists(name) {
		return localeutil.Error("has serialization function %s", name)
	}

	s.items = append(s.items, &Item{
		Name:      name,
		Marshal:   m,
		Unmarshal: u,
	})

	return nil
}

func (s *Serialization) Set(name string, m serializer.MarshalFunc, u serializer.UnmarshalFunc) {
	if item, found := sliceutil.At(s.items, func(i *Item) bool { return name == i.Name }); found {
		item.Marshal = m
		item.Unmarshal = u
		return
	}

	s.items = append(s.items, &Item{
		Name:      name,
		Marshal:   m,
		Unmarshal: u,
	})
}

func (s *Serialization) Exists(name string) bool {
	return sliceutil.Exists(s.items, func(item *Item) bool { return item.Name == name })
}

func (s *Serialization) Delete(name string) {
	s.items = sliceutil.Delete(s.items, func(e *Item) bool { return e.Name == name })
}

func (s *Serialization) Search(name string) (string, serializer.MarshalFunc, serializer.UnmarshalFunc) {
	return s.SearchFunc(func(n string) bool { return n == name })
}

func (s *Serialization) SearchFunc(match func(string) bool) (string, serializer.MarshalFunc, serializer.UnmarshalFunc) {
	if item, ok := sliceutil.At(s.items, func(i *Item) bool { return match(i.Name) }); ok {
		return item.Name, item.Marshal, item.Unmarshal
	}
	return "", nil, nil
}

func (s *Serialization) Len() int { return len(s.items) }

// SPDX-License-Identifier: MIT

// Package mimetype 管理与 Mime type 相关的数据
package mimetypes

import (
	"fmt"
	"strings"

	"github.com/issue9/sliceutil"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"

	"github.com/issue9/web/errs"
	"github.com/issue9/web/internal/header"
)

type (
	Mimetype[M any, U any] struct {
		Name      string
		Problem   string
		Marshal   M
		Unmarshal U
	}

	// Mimetypes 提供对 mimetype 的管理
	//
	// M 表示解码方法地的类型；
	// U 表示编码方法的类型；
	Mimetypes[M any, U any] struct {
		types []*Mimetype[M, U]
	}
)

func New[M any, U any]() *Mimetypes[M, U] {
	return &Mimetypes[M, U]{
		types: make([]*Mimetype[M, U], 0, 10),
	}
}

// Exists 是否存在同名的
func (ms *Mimetypes[M, U]) Exists(name string) bool {
	return sliceutil.Exists(ms.types, func(item *Mimetype[M, U]) bool { return item.Name == name })
}

// Delete 删除指定名称的编码方法
func (ms *Mimetypes[M, U]) Delete(name string) {
	ms.types = sliceutil.Delete(ms.types, func(item *Mimetype[M, U]) bool { return item.Name == name })
}

// Add 添加新的编码方法
//
// name 为编码名称；
// problem 为该编码在返回 [server.Problem] 对象时的 mimetype 报头值，如果为空，则会与 name 值相同；
func (ms *Mimetypes[M, U]) Add(name string, m M, u U, problem string) {
	if ms.Exists(name) {
		panic(fmt.Sprintf("已经存在同名 %s 的编码方法", name))
	}

	if problem == "" {
		problem = name
	}

	ms.types = append(ms.types, &Mimetype[M, U]{
		Name:      name,
		Problem:   problem,
		Marshal:   m,
		Unmarshal: u,
	})
}

// Set 修改或添加指定名称的相关配置
//
// name 用于查找相关的编码方法；
// 如果 problem 为空，则会与 name 值相同；
func (ms *Mimetypes[M, U]) Set(name string, m M, u U, problem string) {
	if problem == "" {
		problem = name
	}

	if item, found := sliceutil.At(ms.types, func(i *Mimetype[M, U]) bool { return name == i.Name }); found {
		item.Marshal = m
		item.Unmarshal = u
		item.Problem = problem
		return
	}

	ms.types = append(ms.types, &Mimetype[M, U]{
		Name:      name,
		Problem:   problem,
		Marshal:   m,
		Unmarshal: u,
	})
}

func (ms *Mimetypes[M, U]) Len() int { return len(ms.types) }

// ContentType 从 content-type 报头中获取解码和字符集函数
//
// h 表示 content-type 报头的内容。如果字符集为 utf-8 或是未指定，返回的字符解码为 nil；
func (ms *Mimetypes[M, U]) ContentType(h string) (U, encoding.Encoding, error) {
	mimetype, charset := header.ParseWithParam(h, "charset")

	item := ms.searchFunc(func(s string) bool { return s == mimetype })
	if item == nil {
		var z U
		return z, nil, errs.NewLocaleError("not found serialization function for %s", mimetype)
	}
	f := item.Unmarshal

	if charset == "" || charset == "utf-8" {
		return f, nil, nil
	}
	e, err := htmlindex.Get(charset)
	if err != nil {
		var z U
		return z, nil, err
	}

	return f, e, nil
}

// MarshalFunc 从 h 解析出当前请求所需要的 mimetype 名称和对应的解码函数
//
// */* 或是空值 表示匹配任意内容，一般会选择第一个元素作匹配；
// xx/* 表示匹配以 xx/ 开头的任意元素，一般会选择 xx/* 开头的第一个元素；
// xx/ 表示完全匹配以 xx/ 的内容
// 如果传递的内容如下：
//
//	application/json;q=0.9,*/*;q=1
//
// 则因为 */* 的 q 值比较高，而返回 */* 匹配的内容
//
// 在不完全匹配的情况下，返回值的名称依然是具体名称。
//
//	text/*;q=0.9
//
// 返回的名称可能是：
//
//	text/plain
func (ms *Mimetypes[M, U]) MarshalFunc(h string) *Mimetype[M, U] {
	if h == "" {
		if item := ms.findMarshal("*/*"); item != nil {
			return item
		}
		return nil
	}

	items := header.ParseQHeader(h, "*/*")
	defer header.PutQHeader(&items)
	for _, item := range items {
		if i := ms.findMarshal(item.Value); i != nil {
			return i
		}
	}

	return nil
}

func (ms *Mimetypes[M, U]) findMarshal(name string) *Mimetype[M, U] {
	switch {
	case ms.Len() == 0:
		return nil
	case name == "" || name == "*/*":
		return ms.searchFunc(func(s string) bool { return true }) // 第一个元素
	case strings.HasSuffix(name, "/*"):
		prefix := name[:len(name)-3]
		return ms.searchFunc(func(s string) bool { return strings.HasPrefix(s, prefix) })
	default:
		return ms.searchFunc(func(s string) bool { return s == name })
	}
}

func (ms *Mimetypes[M, U]) searchFunc(match func(string) bool) *Mimetype[M, U] {
	item, _ := sliceutil.At(ms.types, func(i *Mimetype[M, U]) bool { return match(i.Name) })
	return item
}
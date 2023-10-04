// SPDX-License-Identifier: MIT

// Package mimetypes 管理与 Mimetype 相关的数据
package mimetypes

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/issue9/localeutil"
	"github.com/issue9/sliceutil"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"

	"github.com/issue9/web/internal/header"
)

type (
	Mimetype[M any, U any] struct {
		Name           string
		Problem        string
		MarshalBuilder M
		Unmarshal      U
	}

	// Mimetypes 提供对 mimetype 的管理
	//
	// M 表示解码方法的类型；
	// U 表示编码方法的类型；
	Mimetypes[M any, U any] struct {
		types []*Mimetype[M, U]

		// 根据 types 生成的 Accept 报头
		acceptHeader string
	}
)

func New[M any, U any](cap int) *Mimetypes[M, U] {
	return &Mimetypes[M, U]{types: make([]*Mimetype[M, U], 0, cap)}
}

func (ms *Mimetypes[M, U]) exists(name string) bool {
	return sliceutil.Exists(ms.types, func(item *Mimetype[M, U], _ int) bool { return item.Name == name })
}

// Add 添加新的编码方法
//
// name 为编码名称；
// problem 为该编码在返回 [web.Problem] 对象时的 mimetype 报头值，如果为空，则会与 name 值相同；
func (ms *Mimetypes[M, U]) Add(name string, m M, u U, problem string) {
	if ms.exists(name) {
		panic(fmt.Sprintf("已经存在同名 %s 的编码方法", name))
	}

	if problem == "" {
		problem = name
	}

	ms.types = append(ms.types, &Mimetype[M, U]{
		Name:           name,
		Problem:        problem,
		MarshalBuilder: m,
		Unmarshal:      u,
	})

	names := make([]string, 0, len(ms.types))
	for _, item := range ms.types {
		if !reflect.ValueOf(item.Unmarshal).IsZero() {
			names = append(names, item.Name)
		}
	}
	ms.acceptHeader = strings.Join(names, ",")
}

// ContentType 从请求端提交的 content-type 报头中获取解码和字符集函数
//
// h 表示 content-type 报头的内容。如果字符集为 utf-8 或是未指定，返回的字符解码为 nil；
func (ms *Mimetypes[M, U]) ContentType(h string) (U, encoding.Encoding, error) {
	mimetype, charset := header.ParseWithParam(h, "charset")

	item := ms.searchFunc(func(s string) bool { return s == mimetype })
	if item == nil {
		var z U
		return z, nil, localeutil.Error("not found serialization function for %s", mimetype)
	}
	f := item.Unmarshal

	if charset == "" || charset == header.UTF8Name {
		return f, nil, nil
	}
	e, err := htmlindex.Get(charset)
	if err != nil {
		var z U
		return z, nil, err
	}

	return f, e, nil
}

// Accept 从请求端提交的 accept 报头解析出所需要的解码函数
//
// */* 或是空值 表示匹配任意内容，一般会选择第一个元素作匹配；
// xx/* 表示匹配以 xx/ 开头的任意元素，一般会选择 xx/* 开头的第一个元素；
// xx/ 表示完全匹配以 xx/ 的内容
// 如果传递的内容如下：
//
//	application/json;q=0.9,*/*;q=1
//
// 则因为 */* 的 q 值比较高，而返回 */* 匹配的内容
func (ms *Mimetypes[M, U]) Accept(h string) *Mimetype[M, U] {
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
	case len(ms.types) == 0:
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

func (ms *Mimetypes[M, U]) Search(name string) *Mimetype[M, U] {
	return ms.searchFunc(func(s string) bool { return s == name })
}

func (ms *Mimetypes[M, U]) searchFunc(match func(string) bool) *Mimetype[M, U] {
	item, _ := sliceutil.At(ms.types, func(i *Mimetype[M, U], _ int) bool { return match(i.Name) || match(i.Problem) })
	return item
}

// AcceptHeader 根据当前的内容生成 Accept 报头
func (ms *Mimetypes[M, U]) AcceptHeader() string { return ms.acceptHeader }

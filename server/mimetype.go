// SPDX-License-Identifier: MIT

package server

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/issue9/localeutil"
	"github.com/issue9/sliceutil"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"

	"github.com/issue9/web/internal/header"
)

// ErrUnsupported 返回不支持序列化的错误信息
//
// 当一个对象无法被正常的序列化或是反序列化时返回此错误。
var ErrUnsupported = localeutil.Error("unsupported serialization")

type (
	// MarshalFunc 序列化函数原型
	MarshalFunc func(*Context, any) ([]byte, error)

	// UnmarshalFunc 反序列化函数原型
	UnmarshalFunc func([]byte, any) error

	mimetype struct {
		name      string
		problem   string
		marshal   MarshalFunc
		unmarshal UnmarshalFunc
	}

	Mimetypes struct {
		items []*mimetype
	}
)

func MarshalJSON(ctx *Context, obj any) ([]byte, error) {
	return json.Marshal(obj)
}

func MarshalXML(ctx *Context, obj any) ([]byte, error) {
	return xml.Marshal(obj)
}

// Mimetypes 编解码控制
func (srv *Server) Mimetypes() *Mimetypes { return srv.mimetypes }

func newMimetypes() *Mimetypes {
	return &Mimetypes{
		items: make([]*mimetype, 0, 10),
	}
}

// Exists 是否存在同名的
func (ms *Mimetypes) Exists(name string) bool {
	return sliceutil.Exists(ms.items, func(item *mimetype) bool { return item.name == name })
}

// Delete 删除指定名称的编码方法
func (ms *Mimetypes) Delete(name string) {
	ms.items = sliceutil.Delete(ms.items, func(item *mimetype) bool { return item.name == name })
}

// Add 添加新的编码方法
//
// name 为编码名称；
// problem 为该编码在返回 [Problem] 对象时的 mimetype 报头值，如果为空，则会被赋予 name 相同的值；
func (ms *Mimetypes) Add(name string, m MarshalFunc, u UnmarshalFunc, problem string) {
	if ms.Exists(name) {
		panic(fmt.Sprintf("已经存在同名 %s 的编码方法", name))
	}

	if problem == "" {
		problem = name
	}

	ms.items = append(ms.items, &mimetype{
		name:      name,
		problem:   problem,
		marshal:   m,
		unmarshal: u,
	})
}

// Set 修改指定名称的相关配置
//
// name 用于查找相关的编码方法；
// 如果 problem 为空，会被赋予与 name 相同的值；
func (ms *Mimetypes) Set(name string, m MarshalFunc, u UnmarshalFunc, problem string) {
	if problem == "" {
		problem = name
	}

	if item, found := sliceutil.At(ms.items, func(i *mimetype) bool { return name == i.name }); found {
		item.marshal = m
		item.unmarshal = u
		item.problem = problem
		return
	}

	ms.items = append(ms.items, &mimetype{
		name:      name,
		problem:   problem,
		marshal:   m,
		unmarshal: u,
	})
}

func (ms *Mimetypes) Len() int { return len(ms.items) }

// 从 content-type 报头中获取解码和字符集函数
//
// h 表示 content-type 报头的内容。如果字符集为 utf-8 或是未指定，返回的字符解码为 nil；
func (ms *Mimetypes) contentType(h string) (UnmarshalFunc, encoding.Encoding, error) {
	mimetype, charset := header.ParseWithParam(h, "charset")

	item := ms.searchFunc(func(s string) bool { return s == mimetype })
	if item == nil {
		return nil, nil, localeutil.Error("not found serialization function for %s", mimetype)
	}
	f := item.unmarshal

	if charset == "" || charset == "utf-8" {
		return f, nil, nil
	}
	e, err := htmlindex.Get(charset)
	if err != nil {
		return nil, nil, err
	}

	return f, e, nil
}

// 从 h 解析出当前请求所需要的 mimetype 名称和对应的解码函数
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
func (ms *Mimetypes) marshalFunc(h string) *mimetype {
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

func (ms *Mimetypes) findMarshal(name string) *mimetype {
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

func (ms *Mimetypes) searchFunc(match func(string) bool) *mimetype {
	item, _ := sliceutil.At(ms.items, func(i *mimetype) bool { return match(i.name) })
	return item
}

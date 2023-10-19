// SPDX-License-Identifier: MIT

package codec

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/issue9/localeutil"
	"github.com/issue9/sliceutil"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"

	"github.com/issue9/web"
	"github.com/issue9/web/internal/header"
)

type mimetype struct {
	name           string
	problem        string
	marshalBuilder web.BuildMarshalFunc
	unmarshal      web.UnmarshalFunc
}

func (m *mimetype) Name(problem bool) string {
	if problem {
		return m.problem
	}
	return m.name
}

func (m *mimetype) MarshalBuilder() web.BuildMarshalFunc { return m.marshalBuilder }

func (e *codec) addMimetype(m *Mimetype) {
	if sliceutil.Exists(e.types, func(item *mimetype, _ int) bool { return item.name == m.Name }) {
		panic(fmt.Sprintf("已经存在同名 %s 的编码方法", m.Name))
	}

	if m.Problem == "" {
		m.Problem = m.Name
	}

	e.types = append(e.types, &mimetype{
		name:           m.Name,
		problem:        m.Problem,
		marshalBuilder: m.MarshalBuilder,
		unmarshal:      m.Unmarshal,
	})

	names := make([]string, 0, len(e.types))
	for _, item := range e.types {
		if !reflect.ValueOf(item.unmarshal).IsZero() {
			names = append(names, item.name)
		}
	}
	e.acceptHeader = strings.Join(names, ",")
}

// ContentType 从请求端提交的 content-type 报头中获取解码和字符集函数
//
// h 表示 content-type 报头的内容。如果字符集为 utf-8 或是未指定，返回的字符解码为 nil；
func (e *codec) ContentType(h string) (web.UnmarshalFunc, encoding.Encoding, error) {
	mimetype, charset := header.ParseWithParam(h, "charset")

	item := e.searchFunc(func(s string) bool { return s == mimetype })
	if item == nil {
		return nil, nil, localeutil.Error("not found serialization function for %s", mimetype)
	}
	f := item.unmarshal

	if charset == "" || charset == header.UTF8Name {
		return f, nil, nil
	}
	c, err := htmlindex.Get(charset)
	if err != nil {
		return nil, nil, err
	}

	return f, c, nil
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
func (e *codec) Accept(h string) web.Accepter {
	if h == "" {
		if item := e.findMarshal("*/*"); item != nil {
			return item
		}
		return nil
	}

	items := header.ParseQHeader(h, "*/*")
	defer header.PutQHeader(&items)
	for _, item := range items {
		if i := e.findMarshal(item.Value); i != nil {
			return i
		}
	}

	return nil
}

func (e *codec) findMarshal(name string) *mimetype {
	switch {
	case len(e.types) == 0:
		return nil
	case name == "" || name == "*/*":
		return e.searchFunc(func(s string) bool { return true }) // 第一个元素
	case strings.HasSuffix(name, "/*"):
		prefix := name[:len(name)-3]
		return e.searchFunc(func(s string) bool { return strings.HasPrefix(s, prefix) })
	default:
		return e.searchFunc(func(s string) bool { return s == name })
	}
}

func (e *codec) searchFunc(match func(string) bool) *mimetype {
	item, _ := sliceutil.At(e.types, func(i *mimetype, _ int) bool { return match(i.name) || match(i.problem) })
	return item
}

// AcceptHeader 根据当前的内容生成 Accept 报头
func (e *codec) AcceptHeader() string { return e.acceptHeader }

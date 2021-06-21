// SPDX-License-Identifier: MIT

// Package content 提供对各类媒体数据的处理
package content

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/issue9/qheader"
	"github.com/issue9/sliceutil"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/message"
)

// DefaultMimetype 默认的媒体类型
//
// 在不能获取输入和输出的媒体类型时， 会采用此值作为其默认值。
//
// 若编码函数中指定该类型的函数，则会使用该编码优先匹配 */* 等格式的请求。
const DefaultMimetype = "application/octet-stream"

var (
	// ErrNotFound 表示未找到指定名称的编解码函数
	//
	// 在 Content.Marshal 和 Content.Unmarshal 中会返回该错误。
	ErrNotFound = errors.New("未找到指定名称的编解码函数")

	// ErrExists 存在相同中名称的编解码函数
	//
	// 在 Content.AddMarshal 和 Content.AddUnmarshal 时如果已经存在相同名称，返回此错误。
	ErrExists = errors.New("已经存在相同名称的编解码函数")
)

type (
	// MarshalFunc 将一个对象转换成 []byte 内容时所采用的接口
	MarshalFunc func(v interface{}) ([]byte, error)

	// UnmarshalFunc 将客户端内容转换成一个对象时所采用的接口
	UnmarshalFunc func([]byte, interface{}) error

	mimetype struct {
		name      string
		marshal   MarshalFunc
		unmarshal UnmarshalFunc
	}

	// Content 管理反馈给用户的数据相应在的处理功能
	Content struct {
		mimetypes []*mimetype
		messages  map[int]*resultMessage
		builder   BuildResultFunc
	}

	resultMessage struct {
		status int
		key    message.Reference
		values []interface{}
	}
)

// New 返回 *Content 实例
func New(builder BuildResultFunc) *Content {
	return &Content{
		mimetypes: make([]*mimetype, 0, 10),
		messages:  make(map[int]*resultMessage, 20),
		builder:   builder,
	}
}

// ConentType 从 content-type 报头解析出需要用到的解码函数
func (c *Content) ConentType(header string) (UnmarshalFunc, encoding.Encoding, error) {
	encName, charsetName, err := ParseContentType(header)
	if err != nil {
		return nil, nil, err
	}

	f, err := c.Unmarshal(encName)
	if err != nil {
		return nil, nil, err
	}

	e, err := htmlindex.Get(charsetName)
	if err != nil {
		return nil, nil, err
	}

	return f, e, nil
}

// Unmarshal 查找指定名称的 UnmarshalFunc
func (c *Content) Unmarshal(name string) (UnmarshalFunc, error) {
	for _, mt := range c.mimetypes {
		if mt.name == name {
			return mt.unmarshal, nil
		}
	}
	return nil, ErrNotFound
}

// Marshal 从 header 解析出当前请求所需要的解 mimetype 名称和对应的解码函数
//
// */* 或是空值 表示匹配任意内容，一般会选择第一个元素作匹配；
// xx/* 表示匹配以 xx/ 开头的任意元素，一般会选择 xx/* 开头的第一个元素；
// xx/ 表示完全匹配以 xx/ 的内容
// 如果传递的内容如下：
//  application/json;q=0.9,*/*;q=1
// 则因为 */* 的 q 值比较高，而返回 */* 匹配的内容
//
// 在不完全匹配的情况下，返回值的名称依然是具体名称。
//  text/*;q=0.9
// 返回的名称可能是：
//  text/plain
func (c *Content) Marshal(header string) (string, MarshalFunc, error) {
	if header == "" {
		if mm := c.findMarshal("*/*"); mm != nil {
			return mm.name, mm.marshal, nil
		}
		return "", nil, ErrNotFound
	}

	accepts := qheader.Parse(header, "*/*")
	for _, accept := range accepts {
		if mm := c.findMarshal(accept.Value); mm != nil {
			return mm.name, mm.marshal, nil
		}
	}

	return "", nil, ErrNotFound
}

// AddMimetype 添加编解码函数
//
// m 和 u 可以为 nil，表示仅作为一个占位符使用，具体处理要在 ServeHTTP 中另作处理。
func (c *Content) AddMimetype(name string, m MarshalFunc, u UnmarshalFunc) error {
	if strings.IndexByte(name, '*') >= 0 {
		panic("name 不是一个有效的 mimetype 名称格式")
	}

	for _, mt := range c.mimetypes {
		if mt.name == name {
			return ErrExists
		}
	}

	c.mimetypes = append(c.mimetypes, &mimetype{
		name:      name,
		marshal:   m,
		unmarshal: u,
	})

	sort.SliceStable(c.mimetypes, func(i, j int) bool {
		if c.mimetypes[i].name == DefaultMimetype {
			return true
		}

		if c.mimetypes[j].name == DefaultMimetype {
			return false
		}

		return c.mimetypes[i].name < c.mimetypes[j].name
	})

	return nil
}

// SetMimetype 修改编解码函数
func (c *Content) SetMimetype(name string, m MarshalFunc, u UnmarshalFunc) error {
	for _, mt := range c.mimetypes {
		if mt.name == name {
			mt.marshal = m
			mt.unmarshal = u
			return nil
		}
	}

	return ErrNotFound
}

// DeleteMimetype 删除指定名称的数据
func (c *Content) DeleteMimetype(name string) {
	size := sliceutil.Delete(c.mimetypes, func(i int) bool {
		return c.mimetypes[i].name == name
	})
	c.mimetypes = c.mimetypes[:size]
}

func (c *Content) findMarshal(name string) *mimetype {
	switch {
	case len(c.mimetypes) == 0:
		return nil
	case name == "" || name == "*/*":
		return c.mimetypes[0] // 由 len(marshals) == 0 确保最少有一个元素
	case strings.HasSuffix(name, "/*"):
		prefix := name[:len(name)-3]
		for _, mt := range c.mimetypes {
			if strings.HasPrefix(mt.name, prefix) {
				return mt
			}
		}
	default:
		for _, mt := range c.mimetypes {
			if mt.name == name {
				return mt
			}
		}
	}
	return nil
}

// Messages 错误信息列表
//
// p 用于返回特定语言的内容。
func (c *Content) Messages(p *message.Printer) map[int]string {
	msgs := make(map[int]string, len(c.messages))
	for code, msg := range c.messages {
		msgs[code] = p.Sprintf(msg.key, msg.values...)
	}
	return msgs
}

// AddMessage 添加一条错误信息
//
// status 指定了该错误代码反馈给客户端的 HTTP 状态码；
func (c *Content) AddMessage(status, code int, key message.Reference, v ...interface{}) {
	if _, found := c.messages[code]; found {
		panic(fmt.Sprintf("重复的消息 ID: %d", code))
	}
	c.messages[code] = &resultMessage{status: status, key: key, values: v}
}

// NewResult 返回 Result 实例
//
// 如果找不到 code 对应的错误信息，则会直接 panic。
func (c *Content) NewResult(p *message.Printer, code int) Result {
	msg, found := c.messages[code]
	if !found {
		panic(fmt.Sprintf("不存在的错误代码: %d", code))
	}

	return c.builder(msg.status, code, p.Sprintf(msg.key, msg.values...))
}

// NewResultWithFields 返回 Result 实例
//
// 如果找不到 code 对应的错误信息，则会直接 panic。
func (c *Content) NewResultWithFields(p *message.Printer, code int, fields Fields) Result {
	rslt := c.NewResult(p, code)

	for k, vals := range fields {
		rslt.Add(k, vals...)
	}

	return rslt
}

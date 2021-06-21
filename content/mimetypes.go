// SPDX-License-Identifier: MIT

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
	// 在 Mimetypes.Marshal 和 Mimetypes.Unmarshal 中会返回该错误。
	ErrNotFound = errors.New("未找到指定名称的编解码函数")

	// ErrExists 存在相同中名称的编解码函数
	//
	// 在 Mimetypes.AddMarshal 和 Mimetypes.AddUnmarshal 时如果已经存在相同名称，返回此错误。
	ErrExists = errors.New("已经存在相同名称的编解码函数")
)

type (
	// MarshalFunc 将一个对象转换成 []byte 内容时所采用的接口
	MarshalFunc func(v interface{}) ([]byte, error)

	// UnmarshalFunc 将客户端内容转换成一个对象时所采用的接口
	UnmarshalFunc func([]byte, interface{}) error

	codec struct {
		name      string
		marshal   MarshalFunc
		unmarshal UnmarshalFunc
	}

	// Mimetypes 管理 mimetype 的增删改查
	Mimetypes struct {
		codecs []*codec

		messages map[int]*resultMessage
		builder  BuildResultFunc
	}

	resultMessage struct {
		status int
		key    message.Reference
		values []interface{}
	}
)

// NewMimetypes 返回 *Mimetypes 实例
func NewMimetypes(builder BuildResultFunc) *Mimetypes {
	return &Mimetypes{
		codecs:   make([]*codec, 0, 10),
		messages: make(map[int]*resultMessage, 20),
		builder:  builder,
	}
}

// ConentType 从 content-type 报头解析出需要用到的解码函数
func (mt *Mimetypes) ConentType(header string) (UnmarshalFunc, encoding.Encoding, error) {
	encName, charsetName, err := ParseContentType(header)
	if err != nil {
		return nil, nil, err
	}

	f, err := mt.Unmarshal(encName)
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
func (mt *Mimetypes) Unmarshal(name string) (UnmarshalFunc, error) {
	for _, c := range mt.codecs {
		if c.name == name {
			return c.unmarshal, nil
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
func (mt *Mimetypes) Marshal(header string) (string, MarshalFunc, error) {
	if header == "" {
		if mm := mt.findMarshal("*/*"); mm != nil {
			return mm.name, mm.marshal, nil
		}
		return "", nil, ErrNotFound
	}

	accepts := qheader.Parse(header, "*/*")
	for _, accept := range accepts {
		if mm := mt.findMarshal(accept.Value); mm != nil {
			return mm.name, mm.marshal, nil
		}
	}

	return "", nil, ErrNotFound
}

// Add 添加编解码函数
//
// m 和 u 可以为 nil，表示仅作为一个占位符使用，具体处理要在 ServeHTTP 中另作处理。
func (mt *Mimetypes) Add(name string, m MarshalFunc, u UnmarshalFunc) error {
	if strings.IndexByte(name, '*') >= 0 {
		panic("name 不是一个有效的 mimetype 名称格式")
	}

	for _, c := range mt.codecs {
		if c.name == name {
			return ErrExists
		}
	}

	mt.codecs = append(mt.codecs, &codec{
		name:      name,
		marshal:   m,
		unmarshal: u,
	})

	sort.SliceStable(mt.codecs, func(i, j int) bool {
		if mt.codecs[i].name == DefaultMimetype {
			return true
		}

		if mt.codecs[j].name == DefaultMimetype {
			return false
		}

		return mt.codecs[i].name < mt.codecs[j].name
	})

	return nil
}

// Set 修改编解码函数
func (mt *Mimetypes) Set(name string, m MarshalFunc, u UnmarshalFunc) error {
	for _, c := range mt.codecs {
		if c.name == name {
			c.marshal = m
			c.unmarshal = u
			return nil
		}
	}

	return ErrNotFound
}

// Delete 删除指定名称的数据
func (mt *Mimetypes) Delete(name string) {
	size := sliceutil.Delete(mt.codecs, func(i int) bool {
		return mt.codecs[i].name == name
	})
	mt.codecs = mt.codecs[:size]
}

func (mt *Mimetypes) findMarshal(name string) *codec {
	switch {
	case len(mt.codecs) == 0:
		return nil
	case name == "" || name == "*/*":
		return mt.codecs[0] // 由 len(marshals) == 0 确保最少有一个元素
	case strings.HasSuffix(name, "/*"):
		prefix := name[:len(name)-3]
		for _, c := range mt.codecs {
			if strings.HasPrefix(c.name, prefix) {
				return c
			}
		}
	default:
		for _, c := range mt.codecs {
			if c.name == name {
				return c
			}
		}
	}
	return nil
}

// Messages 错误信息列表
//
// p 用于返回特定语言的内容。
func (mgr *Mimetypes) Messages(p *message.Printer) map[int]string {
	msgs := make(map[int]string, len(mgr.messages))
	for code, msg := range mgr.messages {
		msgs[code] = p.Sprintf(msg.key, msg.values...)
	}
	return msgs
}

// AddMessage 添加一条错误信息
//
// status 指定了该错误代码反馈给客户端的 HTTP 状态码；
func (mgr *Mimetypes) AddMessage(status, code int, key message.Reference, v ...interface{}) {
	if _, found := mgr.messages[code]; found {
		panic(fmt.Sprintf("重复的消息 ID: %d", code))
	}
	mgr.messages[code] = &resultMessage{status: status, key: key, values: v}
}

// NewResult 返回 Result 实例
//
// 如果找不到 code 对应的错误信息，则会直接 panic。
func (mgr *Mimetypes) NewResult(p *message.Printer, code int) Result {
	msg, found := mgr.messages[code]
	if !found {
		panic(fmt.Sprintf("不存在的错误代码: %d", code))
	}

	return mgr.builder(msg.status, code, p.Sprintf(msg.key, msg.values...))
}

// NewResultWithFields 返回 Result 实例
//
// 如果找不到 code 对应的错误信息，则会直接 panic。
func (mgr *Mimetypes) NewResultWithFields(p *message.Printer, code int, fields Fields) Result {
	rslt := mgr.NewResult(p, code)

	for k, vals := range fields {
		rslt.Add(k, vals...)
	}

	return rslt
}

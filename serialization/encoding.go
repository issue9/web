// SPDX-License-Identifier: MIT

package serialization

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/issue9/qheader"
	"github.com/issue9/sliceutil"
)

// Encodings 压缩功能管理
type Encodings struct {
	// 如果指定了这个值，那么会把错误日志输出到此。
	// 若未指定，则不输出内容。
	errlog *log.Logger

	algorithms []*algorithm // 按添加顺序保存，查找 * 时按添加顺序进行比对。

	ignoreTypePrefix []string // 保存通配符匹配的值列表；
	ignoreTypes      []string // 表示完全匹配的值列表。
	allowAny         bool
}

// EncodingWriter 所有压缩对象实现的接口
type EncodingWriter interface {
	io.WriteCloser
	Reset(io.Writer)
}

type algorithm struct {
	name   string
	writer EncodingWriter
}

// NewEncodings 构建一个支持压缩的中间件
//
// errlog 错误日志的输出通道；
// ignoreTypes 表示不需要进行压缩处理的 mimetype 类型，可以是以下格式：
//  - application/json 具体类型；
//  - text* 表示以 text 开头的所有类型；
// 不能传递 *。
func NewEncodings(errlog *log.Logger, ignoreTypes ...string) *Encodings {
	if errlog == nil {
		panic("参数 errlog 不能为空")
	}

	c := &Encodings{
		algorithms: make([]*algorithm, 0, 4),
		errlog:     errlog,
	}

	c.ignoreTypePrefix = make([]string, 0, len(ignoreTypes))
	c.ignoreTypes = make([]string, 0, len(ignoreTypes))
	if len(ignoreTypes) == 0 {
		c.allowAny = true
	} else {
		for _, typ := range ignoreTypes {
			switch {
			case typ == "*":
				panic("无效的值 *")
			case typ[len(typ)-1] == '*':
				// TODO text/* 和 text* 同时存在时，应该删除 text/*
				c.ignoreTypePrefix = append(c.ignoreTypePrefix, typ[:len(typ)-1])
			default:
				c.ignoreTypes = append(c.ignoreTypes, typ)
			}
		}
	}

	return c
}

func (c *Encodings) add(name string, w EncodingWriter) {
	if name == "" || name == "identity" || name == "*" {
		panic("name 值不能为 identity 和 *")
	}

	if w == nil {
		panic("参数 w 不能为空")
	}

	if sliceutil.Count(c.algorithms, func(i int) bool { return c.algorithms[i].name == name }) > 0 {
		panic(fmt.Sprintf("存在相同名称的函数 %s", name))
	}

	c.algorithms = append(c.algorithms, &algorithm{
		name:   name,
		writer: w, // NOTE: 必须传递非空值，否则在 Close 时会出错
	})
}

// Add 添加压缩算法
//
// 当前用户的 Accept-Encoding 的匹配到 * 时，按添加顺序查找真正的匹配项。
// 不能添加名为 identity 和 * 的算法。
//
// 如果未添加任何算法，则每个请求都相当于是 identity 规则。
//
// 返回值表示是否添加成功，若为 false，则表示已经存在相同名称的对象。
func (c *Encodings) Add(algos map[string]EncodingWriter) {
	for name, algo := range algos {
		c.add(name, algo)
	}
}

// Search 从报头中查找最合适的算法
//
//NOTE: 如果返回的 writer 为空值表示不需要压缩
func (c *Encodings) Search(mimetype, header string) (name string, w EncodingWriter, notAcceptable bool) {
	if len(c.algorithms) == 0 || !c.canCompressed(mimetype) {
		return
	}

	accepts := qheader.Parse(header, "*")
	last := accepts[len(accepts)-1]
	if last.Value == "*" { // * 匹配其他任意未在该请求头字段中列出的编码方式
		if last.Q == 0.0 {
			return "", nil, true
		}

		for _, a := range c.algorithms {
			index := sliceutil.Index(accepts, func(i int) bool {
				return accepts[i].Value == a.name
			})
			if index < 0 {
				return a.name, a.writer, false
			}
		}
		return
	}

	var identity *qheader.Header
	for _, accept := range accepts {
		if accept.Err != nil {
			c.errlog.Println(accept.Err)
			continue
		}

		if accept.Value == "identity" { // 除非 q=0，否则表示总是可以被接受
			identity = accept
		}

		for _, a := range c.algorithms {
			if a.name == accept.Value {
				return a.name, a.writer, false
			}
		}
	}
	if identity != nil && identity.Q > 0 {
		a := c.algorithms[0]
		return a.name, a.writer, false
	}

	return // 没有匹配，表示不需要进行压缩
}

// 调用者需要保证 mimetype 的正确性，不能有参数
func (c *Encodings) canCompressed(mimetype string) bool {
	if c.allowAny {
		return true
	}

	for _, val := range c.ignoreTypes {
		if val == mimetype {
			return false
		}
	}

	for _, prefix := range c.ignoreTypePrefix {
		if strings.HasPrefix(mimetype, prefix) {
			return false
		}
	}

	return true
}

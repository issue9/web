// SPDX-License-Identifier: MIT

// Package encoding 处理 Accept-encoding 报头内容
package encoding

import (
	"fmt"
	"io"
	"strings"

	"github.com/issue9/logs/v4"
	"github.com/issue9/qheader"
	"github.com/issue9/sliceutil"
)

type Encodings struct {
	errlog logs.Logger

	pools []*Pool // 按添加顺序保存，查找 * 时按添加顺序进行比对。

	ignoreTypePrefix []string // 保存通配符匹配的值列表；
	ignoreTypes      []string // 表示完全匹配的值列表。
	allowAny         bool
}

// WriterFunc 将普通的 io.Writer 封装成支持压缩功能的对象
type WriterFunc func(w io.Writer) (WriteCloseRester, error)

// NewEncodings 创建 *Encodings
//
// errlog 处理过程中的错误信息输出通道，如果为空表示忽加这些信息；
// ignoreTypes 表示不需要进行压缩处理的 mimetype 类型，可以是以下格式：
//  - application/json 具体类型；
//  - text* 表示以 text 开头的所有类型；
// 不能传递 *。
func NewEncodings(errlog logs.Logger, ignoreTypes ...string) *Encodings {
	c := &Encodings{
		pools:  make([]*Pool, 0, 4),
		errlog: errlog,
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
				// TODO text/* 和 text* 同时存在时，后者包含了前者所有的情况，应该删除 text/*
				c.ignoreTypePrefix = append(c.ignoreTypePrefix, typ[:len(typ)-1])
			default:
				c.ignoreTypes = append(c.ignoreTypes, typ)
			}
		}
	}

	return c
}

func (c *Encodings) add(name string, f WriterFunc) {
	if name == "" || name == "identity" || name == "*" {
		panic("name 值不能为 identity 和 *")
	}

	if f == nil {
		panic("参数 w 不能为空")
	}

	if sliceutil.Count(c.pools, func(e *Pool) bool { return e.name == name }) > 0 {
		panic(fmt.Sprintf("存在相同名称的函数 %s", name))
	}

	c.pools = append(c.pools, newBuilder(name, f))
}

// Add 添加压缩算法
//
// 当前用户的 Accept-Encoding 的匹配到 * 时，按添加顺序查找真正的匹配项。
// 不能添加名为 identity 和 * 的算法。
//
// 如果未添加任何算法，则每个请求都相当于是 identity 规则。
//
// 返回值表示是否添加成功，若为 false，则表示已经存在相同名称的对象。
func (c *Encodings) Add(algos map[string]WriterFunc) {
	for name, algo := range algos {
		c.add(name, algo)
	}
}

// Search 从报头中查找最合适的算法
//
// 如果返回的 w 为空值表示不需要压缩。
func (c *Encodings) Search(mimetype, header string) (w *Pool, notAcceptable bool) {
	if len(c.pools) == 0 || !c.canCompressed(mimetype) {
		return
	}

	accepts := qheader.Parse(header, "*")
	if accepts == nil || len(accepts.Items) == 0 {
		return
	}

	last := accepts.Items[len(accepts.Items)-1]
	if last.Value == "*" { // * 匹配其他任意未在该请求头字段中列出的编码方式
		if last.Q == 0.0 {
			return nil, true
		}

		for _, a := range c.pools {
			index := sliceutil.Index(accepts.Items, func(e *qheader.Item) bool {
				return e.Value == a.name
			})
			if index < 0 {
				return a, false
			}
		}
		return
	}

	var identity *qheader.Item
	for _, accept := range accepts.Items {
		if accept.Err != nil {
			if c.errlog != nil {
				c.errlog.Error(accept.Err)
			}
			continue
		}

		if accept.Value == "identity" { // 除非 q=0，否则表示总是可以被接受
			identity = accept
		}

		for _, a := range c.pools {
			if a.name == accept.Value {
				return a, false
			}
		}
	}
	if identity != nil && identity.Q > 0 {
		a := c.pools[0]
		return a, false
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

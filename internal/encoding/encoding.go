// SPDX-License-Identifier: MIT

// Package encoding 处理 Accept-encoding 报头内容
package encoding

import (
	"strings"

	"github.com/issue9/sliceutil"

	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/logs"
)

type Encodings struct {
	errlog logs.Logger
	algs   []*Alg
}

func NewEncodings(errlog logs.Logger) *Encodings {
	return &Encodings{
		algs:   make([]*Alg, 0, 10),
		errlog: errlog,
	}
}

// Add 添加一种压缩算法
//
// name 算法名称，可以重复；
func (c *Encodings) Add(name string, f NewEncodingFunc, contentType ...string) {
	if name == "" || name == "identity" || name == "*" {
		panic("name 值不能为 identity 和 *")
	}

	if f == nil {
		panic("参数 f 不能为空")
	}

	c.algs = append(c.algs, newAlg(name, f, contentType...))
}

// Search 从报头中查找最合适的算法
//
// 如果返回的 w 为空值表示不需要压缩。
// 当有多个符合时，按添加顺序拿第一个符合条件数据。
func (c *Encodings) Search(contentType, h string) (w *Alg, notAcceptable bool) {
	if len(c.algs) == 0 {
		return
	}

	accepts := header.ParseQHeader(h, "*")
	defer header.PutQHeader(&accepts)
	if len(accepts) == 0 {
		return
	}

	pools := c.getMatchAlgs(contentType)
	if len(pools) == 0 {
		return
	}

	if last := accepts[len(accepts)-1]; last.Value == "*" { // * 匹配其他任意未在该请求头字段中列出的编码方式
		if last.Q == 0.0 {
			return nil, true
		}

		for _, p := range pools {
			exists := sliceutil.Exists(accepts, func(e *header.Item) bool {
				return e.Value == p.name
			})
			if !exists {
				return p, false
			}
		}
		return
	}

	var identity *header.Item
	for _, accept := range accepts {
		if accept.Err != nil {
			if c.errlog != nil {
				c.errlog.Error(accept.Err)
			}
			continue
		}

		if accept.Value == "identity" { // 除非 q=0，否则表示总是可以被接受
			identity = accept
		}

		for _, a := range pools {
			if a.name == accept.Value {
				return a, false
			}
		}
	}
	if identity != nil && identity.Q > 0 {
		a := pools[0]
		return a, false
	}

	return // 没有匹配，表示不需要进行压缩
}

func (c *Encodings) getMatchAlgs(contentType string) []*Alg {
	algs := make([]*Alg, 0, len(c.algs))

LOOP:
	for _, alg := range c.algs {
		for _, s := range alg.allowTypes {
			if s == contentType {
				algs = append(algs, alg)
				continue LOOP
			}
		}

		for _, p := range alg.allowTypesPrefix {
			if strings.HasPrefix(contentType, p) {
				algs = append(algs, alg)
				continue LOOP
			}
		}
	}

	return algs
}

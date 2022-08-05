// SPDX-License-Identifier: MIT

// Package encoding 处理 Accept-encoding 报头内容
package encoding

import (
	"fmt"
	"sort"
	"strings"

	"github.com/issue9/logs/v4"
	"github.com/issue9/sliceutil"

	"github.com/issue9/web/internal/header"
)

type Encodings struct {
	errlog logs.Logger

	pools map[string]*Pool // 按添加顺序保存，查找 * 时按添加顺序进行比对。

	// contentType 是具体值的，比如 text/xml
	allowTypes map[string][]*Pool

	// contentType 是模糊类型的，比如 text/*，
	// 只有在 allowTypes 找不到时，才在此处查找。
	allowTypesPrefix []prefix
}

type prefix struct {
	prefix string
	pools  []*Pool
}

func NewEncodings(errlog logs.Logger) *Encodings {
	return &Encodings{
		pools:  make(map[string]*Pool, 10),
		errlog: errlog,

		allowTypes:       make(map[string][]*Pool, 10),
		allowTypesPrefix: make([]prefix, 0, 10),
	}
}

// Allow 允许 contentType 采用的压缩方式
//
// id 是指由 Add 中指定的值；
// contentType 表示经由 Accept-Encoding 提交的值，该值不能是 identity 和 *；
//
// 如果未添加任何算法，则每个请求都相当于是 identity 规则。
func (c *Encodings) Allow(contentType string, id ...string) {
	if len(id) == 0 {
		panic("id 不能为空")
	}

	pools := make([]*Pool, 0, len(id))
	for _, i := range id {
		p, found := c.pools[i]
		if !found {
			panic(fmt.Sprintf("未找到 id 为 %s 表示的算法", i))
		}
		pools = append(pools, p)
	}
	if indexes := sliceutil.Dup(pools, func(i, j *Pool) bool { return i.name == j.name }); len(indexes) > 0 {
		panic(fmt.Sprintf("id 引用中存在多个名为 %s 的算法", pools[indexes[0]].name))
	}

	switch {
	case contentType[len(contentType)-1] == '*':
		p := contentType[:len(contentType)-1]
		if sliceutil.Exists(c.allowTypesPrefix, func(e prefix) bool { return e.prefix == p }) {
			panic(fmt.Sprintf("已经存在对 %s 的压缩规则", contentType))
		}

		c.allowTypesPrefix = append(c.allowTypesPrefix, prefix{pools: pools, prefix: p})
		// 按 prefix 从长到短排序
		sort.SliceStable(c.allowTypesPrefix, func(i, j int) bool {
			return len(c.allowTypesPrefix[i].prefix) > len(c.allowTypesPrefix[j].prefix)
		})
	default:
		if _, found := c.allowTypes[contentType]; found {
			panic(fmt.Sprintf("已经存在对 %s 的压缩规则", contentType))
		}
		c.allowTypes[contentType] = pools
	}
}

// Add 添加压缩算法
//
// id 表示当前算法的唯一名称，在 Allow 中可以用来查找使用；
// name 表示通过 Accept-Encoding 匹配的名称；
// f 表示生成压缩对象的方法；
func (c *Encodings) Add(id, name string, f NewEncodingFunc) {
	if name == "" || name == "identity" || name == "*" {
		panic("name 值不能为 identity 和 *")
	}

	if f == nil {
		panic("参数 w 不能为空")
	}

	if _, found := c.pools[id]; found {
		panic(fmt.Sprintf("存在相同 ID %s 的函数", id))
	}
	c.pools[id] = newPool(name, f)
}

// Search 从报头中查找最合适的算法
//
// 如果返回的 w 为空值表示不需要压缩。
func (c *Encodings) Search(contentType, h string) (w *Pool, notAcceptable bool) {
	if len(c.pools) == 0 {
		return
	}

	accepts := header.ParseQHeader(h, "*")
	defer header.PutQHeader(&accepts)
	if len(accepts) == 0 {
		return
	}

	pools := c.getPools(contentType)
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

// 调用者需要保证 mimetype 的正确性，不能有参数
func (c *Encodings) getPools(contentType string) []*Pool {
	for t, pools := range c.allowTypes {
		if t == contentType {
			return pools
		}
	}

	for _, p := range c.allowTypesPrefix {
		if strings.HasPrefix(contentType, p.prefix) {
			return p.pools
		}
	}

	return nil
}

// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

// Package qheader 与报头相关的处理方法
package qheader

import (
	"cmp"
	"slices"
	"strconv"
	"strings"
	"sync"
)

var itemsPool = &sync.Pool{New: func() any { return &[]Item{} }}

// Item 表示报头内容的单个元素内容
//
// 比如 zh-cmt;q=0.8, zh-cmn;q=1, 拆分成两个 Item 对象。
type Item struct {
	Value string
	Q     float32
}

func FreeItems(items *[]Item) { itemsPool.Put(items) }

// ParseQHeader 解析报头内容
//
// 排序方式如下:
//
// Q 值大的靠前，如果 Q 值相同，则全名的比带通配符的靠前，*/* 最后，都是全名则按原来顺序返回。
//
// header 表示报头的内容，如果为空则相当于支持所有值；
// any 表示通配符的值，只能是 */*、* 和空值，其它情况则 panic；
//
// allow 为报头解析出来的正常项；
// deny 为报头解析出来之后 Q 值为零的项的值；
// errs 为报头中无法正常解析的原始项；
func ParseQHeader(header string, any string) (allow []Item, deny, errs []string) {
	if any != "*" && any != "*/*" && any != "" {
		panic("any 值错误")
	}

	if header == "" {
		header = any
	}

	headers := strings.Split(header, ",")
	allow = *itemsPool.Get().(*[]Item)
	allow = allow[:0]

	for _, h := range headers {
		if h = strings.TrimSpace(h); h == "" {
			continue
		}

		q := 1.0
		v, p := ParseWithParam(h, "q")
		if p != "" {
			var err error
			if q, err = strconv.ParseFloat(p, 32); err != nil {
				errs = append(errs, h)
				continue
			}
		}

		if q <= 0 {
			deny = append(deny, v)
		} else {
			allow = append(allow, Item{Value: v, Q: float32(q)})
		}
	}

	sortItems(allow, any)

	return allow, deny, errs
}

func sortItems(items []Item, any string) {
	slices.SortStableFunc(items, func(i, j Item) int {
		if i.Q != j.Q {
			return cmp.Compare(j.Q, i.Q)
		}

		switch {
		case i.Value == any:
			return 1
		case j.Value == any:
			return -1
		case i.hasWildcard(): // 如果 any == * 则此判断不启作用
			return 1
		case j.hasWildcard():
			return -1
		default:
			return 0
		}
	})
}

func (header *Item) hasWildcard() bool { return strings.HasSuffix(header.Value, "/*") }

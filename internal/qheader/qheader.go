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

var itemsPool = &sync.Pool{New: func() any { return []*Item{} }}

// Item 表示报头内容的单个元素内容
//
// 比如 zh-cmt;q=0.8, zh-cmn;q=1, 拆分成两个 Item 对象。
type Item struct {
	Value string
	Q     float32
}

func PutQHeader(items []*Item) { itemsPool.Put(items) }

// ParseQHeader 解析报头内容
//
// 排序方式如下:
//
// Q 值大的靠前，如果 Q 值相同，则全名的比带通配符的靠前，*/* 最后，都是全名则按原来顺序返回。
//
// header 表示报头的内容；
// any 表示通配符的值，只能是 */*、* 和空值，其它情况则 panic；
//
// 所有 Item.Q 为零的会被忽略，如果解析 q 出错会返回错误；
func ParseQHeader(header string, any string) (items []*Item, err error) {
	if any != "*" && any != "*/*" && any != "" {
		panic("any 值错误")
	}

	headers := strings.Split(header, ",")
	items = itemsPool.Get().([]*Item)
	il, hl := len(items), len(headers)
	if il > hl {
		items = items[:hl]
	} else if il < hl {
		for i := 0; i < hl-il; i++ {
			items = append(items, &Item{})
		}
	}

	count := 0
	for _, h := range headers {
		if h = strings.TrimSpace(h); h == "" {
			continue
		}

		q := 1.0
		var err error
		v, p := ParseWithParam(h, "q")
		if p != "" {
			if q, err = strconv.ParseFloat(p, 32); err != nil {
				return nil, err
			}
		}

		// 如果是 *;q=0 这样特殊的项需要保留
		if q == 0 && v != any {
			continue
		}

		// NOTE: 从 pool 取得的值，需要全部覆盖。
		item := items[count]
		item.Value = v
		item.Q = float32(q)

		count++
	}
	items = items[:count]
	sortItems(items, any)
	return items, nil
}

func sortItems(items []*Item, any string) {
	slices.SortStableFunc(items, func(i, j *Item) int {
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
			return 1
		}
	})
}

func (header *Item) hasWildcard() bool { return strings.HasSuffix(header.Value, "/*") }

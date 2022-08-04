// SPDX-License-Identifier: MIT

package header

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

var itemsPool = &sync.Pool{New: func() any { return new([]*Item) }}

// Item 表示报头内容的单个元素内容
//
// 比如 zh-cmt;q=0.8, zh-cmn;q=1, 拆分成两个 Item 对象。
type Item struct {
	Value string
	Q     float32
	Err   error // 如果 Q 解析出错会出现在此
}

func PutQHeader(items *[]*Item) { itemsPool.Put(items) }

// ParseQHeader 解析报头内容
//
// 排序方式如下:
//
// Q 值大的靠前，如果 Q 值相同，则全名的比带通配符的靠前，*/* 最后，都是全名则按原来顺序返回。
//
// header 表示报头的内容；
// any 表示通配符的值，只能是 */*、* 和空值，其它情况则 panic；
func ParseQHeader(header string, any string) (items []*Item) {
	if any != "*" && any != "*/*" && any != "" {
		panic("any 值错误")
	}

	headers := strings.Split(header, ",")
	items = *itemsPool.Get().(*[]*Item)
	il, hl := len(items), len(headers)
	if il > hl {
		items = items[:hl]
	} else if il < hl {
		for i := 0; i < hl-il; i++ {
			items = append(items, &Item{})
		}
	}

	count := 0
	for index, h := range headers {
		if h = strings.TrimSpace(h); h == "" {
			continue
		}

		v, p := ParseWithParam(h, "q")
		q := 1.0
		var err error
		if p != "" {
			q, err = strconv.ParseFloat(p, 32)
			if err != nil {
				q = 0
			}
		}

		count++
		item := items[index]
		item.Value = v
		item.Q = float32(q)
		item.Err = err
	}
	items = items[:count]
	sortItems(items, any)
	return items
}

func sortItems(items []*Item, any string) {
	sort.SliceStable(items, func(i, j int) bool {
		ii := items[i]
		jj := items[j]

		if ii.Err != nil {
			return false
		}
		if jj.Err != nil {
			return true
		}

		if ii.Q != jj.Q {
			return ii.Q > jj.Q
		}

		switch {
		case ii.Value == any:
			return false
		case jj.Value == any:
			return true
		case ii.hasWildcard(): // 如果 any == * 则此判断不启作用
			return false
		case jj.hasWildcard():
			return true
		default:
			return false
		}
	})
}

func (header *Item) hasWildcard() bool { return strings.HasSuffix(header.Value, "/*") }

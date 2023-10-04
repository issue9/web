// SPDX-License-Identifier: MIT

package compress

import (
	"io"
	"strings"

	"github.com/issue9/localeutil"
	"github.com/issue9/sliceutil"

	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/logs"
)

// Compresses 压缩算法的管理
type Compresses struct {
	compresses   []*NamedCompress
	acceptHeader string

	disable bool // 是否禁用压缩输出
}

func NewCompresses(cap int, disable bool) *Compresses {
	return &Compresses{
		compresses: make([]*NamedCompress, 0, cap),
		disable:    disable,
	}
}

// ContentEncoding 根据 Content-Encoding 报头返回相应的解码对象
//
// name 编码名称，即 Content-Encoding 报头内容；
// r 为未解码的内容；
func (e *Compresses) ContentEncoding(name string, r io.Reader) (io.ReadCloser, error) {
	if c, f := sliceutil.At(e.compresses, func(item *NamedCompress, _ int) bool { return item.Name() == name }); f {
		return c.Compress().Decoder(r)
	}
	return nil, localeutil.Error("not found compress for %s", name)
}

// AcceptEncoding 根据客户端的 Accept-Encoding 报头查找最合适的算法
//
// 如果返回的 w 为空值表示不需要压缩。
// 当有多个符合时，按添加顺序拿第一个符合条件数据。
// l 表示解析报头过程中的错误信息，可以为空，表示不输出信息；
func (e *Compresses) AcceptEncoding(contentType, h string, l logs.Logger) (w *NamedCompress, notAcceptable bool) {
	if len(e.compresses) == 0 || e.IsDisable() {
		return
	}

	accepts := header.ParseQHeader(h, "*")
	defer header.PutQHeader(&accepts)
	if len(accepts) == 0 {
		return
	}

	indexes := e.getMatchCompresses(contentType)
	if len(indexes) == 0 {
		return
	}

	if last := accepts[len(accepts)-1]; last.Value == "*" { // * 匹配其他任意未在该请求头字段中列出的编码方式
		if last.Q == 0.0 {
			return nil, true
		}

		for _, index := range indexes {
			curr := e.compresses[index]
			if !sliceutil.Exists(accepts, func(i *header.Item, _ int) bool { return i.Value == curr.name }) {
				return curr, false
			}
		}
		return
	}

	var identity *header.Item
	for _, accept := range accepts {
		if accept.Err != nil && l != nil {
			l.Error(accept.Err)
			continue
		}

		if accept.Value == "identity" { // 除非 q=0，否则表示总是可以被接受
			identity = accept
		}

		for _, index := range indexes {
			if curr := e.compresses[index]; curr.name == accept.Value {
				return curr, false
			}
		}
	}
	if identity != nil && identity.Q > 0 {
		return e.compresses[indexes[0]], false
	}

	return // 没有匹配，表示不需要进行压缩
}

func (e *Compresses) getMatchCompresses(contentType string) []int {
	indexes := make([]int, 0, len(e.compresses))

LOOP:
	for index, c := range e.compresses {
		for _, s := range c.allowTypes {
			if s == contentType {
				indexes = append(indexes, index)
				continue LOOP
			}
		}

		for _, p := range c.allowTypesPrefix {
			if strings.HasPrefix(contentType, p) {
				indexes = append(indexes, index)
				continue LOOP
			}
		}
	}

	return indexes
}

// AcceptEncodingHeader 生成 AcceptEncoding 报头内容
func (e *Compresses) AcceptEncodingHeader() string { return e.acceptHeader }

// Add 添加新的压缩算法
func (e *Compresses) Add(name string, c Compressor, ct ...string) *Compresses {
	types := make([]string, 0, len(ct))
	prefix := make([]string, 0, len(ct))
	for _, c := range ct {
		if c == "" {
			continue
		}

		if c[len(c)-1] == '*' {
			prefix = append(prefix, c[:len(c)-1])
		} else {
			types = append(types, c)
		}
	}

	e.compresses = append(e.compresses, &NamedCompress{
		name:             name,
		compress:         c,
		allowTypes:       types,
		allowTypesPrefix: prefix,
	})

	names := make([]string, 0, len(e.compresses))
	for _, item := range e.compresses {
		names = append(names, item.name)
	}
	names = sliceutil.Unique(names, func(i, j string) bool { return i == j })
	e.acceptHeader = strings.Join(names, ",")

	return e
}

func (e *Compresses) SetDisable(disable bool) { e.disable = disable }

func (e *Compresses) IsDisable() bool { return e.disable }

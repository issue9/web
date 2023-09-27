// SPDX-License-Identifier: MIT

package compress

import (
	"strings"

	"github.com/issue9/sliceutil"

	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/logs"
)

type Compresses struct {
	compresses   []*NamedCompress
	acceptHeader string
}

func NewCompresses(cap int) *Compresses {
	return &Compresses{compresses: make([]*NamedCompress, 0, cap)}
}

func (e *Compresses) ContentEncoding() {
	// TODO
}

// AcceptEncoding 根据客户端的 AcceptEncoding 报头查找最合适的算法
//
// 如果返回的 w 为空值表示不需要压缩。
// 当有多个符合时，按添加顺序拿第一个符合条件数据。
// l 表示解析报头过程中的错误信息，可以为空，表示不输出信息；
func (e *Compresses) AcceptEncoding(contentType, h string, l logs.Logger) (w *NamedCompress, notAcceptable bool) {
	if len(e.compresses) == 0 {
		return
	}

	accepts := header.ParseQHeader(h, "*")
	defer header.PutQHeader(&accepts)
	if len(accepts) == 0 {
		return
	}

	pools := e.getMatchAlgs(contentType)
	if len(pools) == 0 {
		return
	}

	if last := accepts[len(accepts)-1]; last.Value == "*" { // * 匹配其他任意未在该请求头字段中列出的编码方式
		if last.Q == 0.0 {
			return nil, true
		}

		for _, p := range pools {
			exists := sliceutil.Exists(accepts, func(e *header.Item, _ int) bool {
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
		if accept.Err != nil && l != nil {
			l.Error(accept.Err)
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

func (e *Compresses) getMatchAlgs(contentType string) []*NamedCompress {
	algs := make([]*NamedCompress, 0, len(e.compresses))

LOOP:
	for _, alg := range e.compresses {
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

// AcceptEncodingHeader 生成 AcceptEncoding 报头内容
func (e *Compresses) AcceptEncodingHeader() string { return e.acceptHeader }

func (e *Compresses) Add(name string, c Compress, ct ...string) *Compresses {
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

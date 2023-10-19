// SPDX-License-Identifier: MIT

package codec

import (
	"io"
	"strings"

	"github.com/issue9/localeutil"
	"github.com/issue9/sliceutil"

	"github.com/issue9/web"
	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/logs"
)

type namedCompressor struct {
	name     string
	compress Compressor

	// contentType 是具体值的，比如 text/xml
	allowTypes []string

	// contentType 是模糊类型的，比如 text/*，
	// 只有在 allowTypes 找不到时，才在此处查找。
	allowTypesPrefix []string
}

// AddCompressor 添加新的压缩算法
func (e *Codec) AddCompressor(name string, c Compressor, ct ...string) *Codec {
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

	e.compresses = append(e.compresses, &namedCompressor{
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
	e.acceptEncodingHeader = strings.Join(names, ",")

	return e
}

func (e *Codec) ContentEncoding(name string, r io.Reader) (io.ReadCloser, error) {
	if c, f := sliceutil.At(e.compresses, func(item *namedCompressor, _ int) bool { return item.name == name }); f {
		return c.compress.NewDecoder(r)
	}
	return nil, localeutil.Error("not found compress for %s", name)
}

func (e *Codec) AcceptEncoding(contentType, h string, l logs.Logger) (c web.CompressorWriterFunc, name string, notAcceptable bool) {
	if len(e.compresses) == 0 || !e.CanCompress() {
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
			return nil, "", true
		}

		for _, index := range indexes {
			curr := e.compresses[index]
			if !sliceutil.Exists(accepts, func(i *header.Item, _ int) bool { return i.Value == curr.name }) {
				return curr.compress.NewEncoder, curr.name, false
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

		if accept.Value == header.Identity { // 除非 q=0，否则表示总是可以被接受
			identity = accept
		}

		for _, index := range indexes {
			if curr := e.compresses[index]; curr.name == accept.Value {
				return curr.compress.NewEncoder, curr.name, false
			}
		}
	}
	if identity != nil && identity.Q > 0 {
		c := e.compresses[indexes[0]]
		return c.compress.NewEncoder, c.name, false
	}

	return // 没有匹配，表示不需要进行压缩
}

func (e *Codec) getMatchCompresses(contentType string) []int {
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
func (e *Codec) AcceptEncodingHeader() string { return e.acceptEncodingHeader }

func (e *Codec) EnableCompress() { e.disableCompress = false }

func (e *Codec) DisableCompress() { e.disableCompress = true }

func (e *Codec) CanCompress() bool { return !e.disableCompress }

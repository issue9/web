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

func (e *codec) addCompression(c *Compression) {
	cc := *c // 复制，防止通过配置项修改内容。
	e.compressions = append(e.compressions, &cc)

	names := make([]string, 0, len(e.compressions))
	for _, item := range e.compressions {
		names = append(names, item.Compressor.Name())
	}
	names = sliceutil.Unique(names, func(i, j string) bool { return i == j })
	e.acceptEncodingHeader = strings.Join(names, ",")
}

func (e *codec) ContentEncoding(name string, r io.Reader) (io.ReadCloser, error) {
	if name == "" {
		return io.NopCloser(r), nil
	}

	if c, f := sliceutil.At(e.compressions, func(item *Compression, _ int) bool { return item.Compressor.Name() == name }); f {
		return c.Compressor.NewDecoder(r)
	}
	return nil, localeutil.Error("not found compress for %s", name)
}

func (e *codec) AcceptEncoding(contentType, h string, l logs.Logger) (c web.CompressorWriterFunc, name string, notAcceptable bool) {
	if len(e.compressions) == 0 || !e.CanCompress() {
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
			curr := e.compressions[index]
			if !sliceutil.Exists(accepts, func(i *header.Item, _ int) bool { return i.Value == curr.Compressor.Name() }) {
				return curr.Compressor.NewEncoder, curr.Compressor.Name(), false
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
			if curr := e.compressions[index]; curr.Compressor.Name() == accept.Value {
				return curr.Compressor.NewEncoder, curr.Compressor.Name(), false
			}
		}
	}
	if identity != nil && identity.Q > 0 {
		c := e.compressions[indexes[0]]
		return c.Compressor.NewEncoder, c.Compressor.Name(), false
	}

	return // 没有匹配，表示不需要进行压缩
}

func (e *codec) getMatchCompresses(contentType string) []int {
	indexes := make([]int, 0, len(e.compressions))

LOOP:
	for index, c := range e.compressions {
		if c.wildcard {
			indexes = append(indexes, index)
			continue
		}

		for _, s := range c.Types {
			if s == contentType {
				indexes = append(indexes, index)
				continue LOOP
			}
		}

		for _, p := range c.wildcardSuffix {
			if strings.HasPrefix(contentType, p) {
				indexes = append(indexes, index)
				continue LOOP
			}
		}
	}

	return indexes
}

// AcceptEncodingHeader 生成 AcceptEncoding 报头内容
func (e *codec) AcceptEncodingHeader() string { return e.acceptEncodingHeader }

func (e *codec) SetCompress(enable bool) { e.disableCompress = !enable }

func (e *codec) CanCompress() bool { return !e.disableCompress }

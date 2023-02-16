// SPDX-License-Identifier: MIT

package encoding

import (
	"io"
	"sync"
)

var algWriterPool = sync.Pool{New: func() any { return &algWriter{} }}

type Alg struct {
	name string     // 算法名称
	pool *sync.Pool // 算法的对象池

	// contentType 是具体值的，比如 text/xml
	allowTypes []string

	// contentType 是模糊类型的，比如 text/*，
	// 只有在 allowTypes 找不到时，才在此处查找。
	allowTypesPrefix []string
}

// 当调用 algWriter.Close 时自动回收到 Pool 中
type algWriter struct {
	WriteCloseRester
	b *Alg
}

func (e *algWriter) Close() error {
	err := e.WriteCloseRester.Close()
	e.b.pool.Put(e.WriteCloseRester)
	algWriterPool.Put(e)
	return err
}

func newAlg(name string, f NewEncodingFunc, ct ...string) *Alg {
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
	return &Alg{
		name: name,
		pool: &sync.Pool{New: func() any { return f() }},

		allowTypes:       types,
		allowTypesPrefix: prefix,
	}
}

func (p *Alg) Get(w io.Writer) io.WriteCloser {
	ww := p.pool.Get().(WriteCloseRester)
	ww.Reset(w)

	pw := algWriterPool.Get().(*algWriter)
	pw.b = p
	pw.WriteCloseRester = ww
	return pw
}

func (p *Alg) Name() string { return p.name }

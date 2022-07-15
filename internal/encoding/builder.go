// SPDX-License-Identifier: MIT

package encoding

import (
	"bytes"
	"io"
	"sync"
)

type Builder struct {
	name string
	pool *sync.Pool
}

type encodingW struct {
	WriteCloseRester
	b *Builder
}

func (e *encodingW) Close() error {
	err := e.WriteCloseRester.Close()
	e.b.pool.Put(e.WriteCloseRester)
	return err
}

// WriterFunc 将普通的 io.Writer 封装成支持压缩功能的对象
type WriterFunc func(w io.Writer) (WriteCloseRester, error)

func newBuilder(name string, f WriterFunc) *Builder {
	return &Builder{
		name: name,
		pool: &sync.Pool{New: func() any {
			w, err := f(&bytes.Buffer{}) // NOTE: 必须传递非空值，否则在 Close 时会出错
			if err != nil {
				panic(err)
			}
			return w
		}},
	}
}

func (b *Builder) Build(w io.Writer) io.WriteCloser {
	ww := b.pool.Get().(WriteCloseRester)
	ww.Reset(w)
	return &encodingW{b: b, WriteCloseRester: ww}
}

func (b *Builder) Name() string { return b.name }

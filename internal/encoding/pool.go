// SPDX-License-Identifier: MIT

package encoding

import (
	"bytes"
	"io"
	"sync"
)

var poolWriterPool = sync.Pool{New: func() any { return &poolWriter{} }}

// Pool 压缩对象池
//
// 每个 Pool 对象与特定的压缩对象关联，可以复用这些压缩对象。
type Pool struct {
	name string
	pool *sync.Pool
}

// 当调用 poolWriter.Close 时自动回收到 Pool 中
type poolWriter struct {
	WriteCloseRester
	b *Pool
}

func (e *poolWriter) Close() error {
	err := e.WriteCloseRester.Close()
	e.b.pool.Put(e.WriteCloseRester)
	poolWriterPool.Put(e)
	return err
}

func newPool(name string, f WriterFunc) *Pool {
	return &Pool{
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

func (b *Pool) Get(w io.Writer) io.WriteCloser {
	ww := b.pool.Get().(WriteCloseRester)
	ww.Reset(w)
	pw := poolWriterPool.Get().(*poolWriter)
	pw.b = b
	pw.WriteCloseRester = ww
	return pw
}

func (b *Pool) Name() string { return b.name }

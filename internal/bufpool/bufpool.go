// SPDX-License-Identifier: MIT

// Package bufpool 提供 bytes.Buffer 的对象池
package bufpool

import (
	"bytes"
	"sync"
)

const bufMaxSize = 1024

var bufPool = &sync.Pool{New: func() any { return &bytes.Buffer{} }}

func New() *bytes.Buffer {
	b := bufPool.Get().(*bytes.Buffer)
	b.Reset()
	return b
}

func Put(p *bytes.Buffer) {
	if p.Cap() < bufMaxSize {
		bufPool.Put(p)
	}
}

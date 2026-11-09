// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

// Package bufpool 提供 [bytes.Buffer] 的对象池
package bufpool

import (
	"bytes"
	"sync"
)

var bufPool = &sync.Pool{New: func() any { return &bytes.Buffer{} }}

// New 声明缓存的 [bytes.Buffer] 对象
func New() *bytes.Buffer {
	b := bufPool.Get().(*bytes.Buffer)
	b.Reset()
	return b
}

func Put(p *bytes.Buffer) {
	if p.Cap() < 1024 {
		bufPool.Put(p)
	}
}

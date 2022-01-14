// SPDX-License-Identifier: MIT

package filesystem

import (
	"io/fs"
)

type MultipleFS struct {
	f []fs.FS
}

// NewMultipleFS 将多个 fs.FS 合并成一个 fs.FS 对象
//
// 每次查找时，都顺序查找第一个存在的对象并返回。
func NewMultipleFS(f ...fs.FS) *MultipleFS { return &MultipleFS{f: f} }

// Add 添加另一个文件系统
func (f *MultipleFS) Add(fsys ...fs.FS) { f.f = append(f.f, fsys...) }

func (f *MultipleFS) Open(name string) (fs.File, error) {
	for _, fsys := range f.f {
		if existsFS(fsys, name) {
			return fsys.Open(name)
		}
	}

	return nil, fs.ErrNotExist
}

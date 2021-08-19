// SPDX-License-Identifier: MIT

package filesystem

import "io/fs"

type MultipleFS struct {
	f []fs.FS
}

// MultipleFS 将多个 fs.FS 合并成一个 fs.FS 对象
//
// 每次查找时，都顺序查找第一个存在的对象并返回。
func NewMultipleFS(f ...fs.FS) *MultipleFS {
	return &MultipleFS{f: f}
}

func (f *MultipleFS) Add(fsys ...fs.FS) {
	f.f = append(f.f, fsys...)
}

func (f *MultipleFS) Open(name string) (fs.File, error) {
	for _, fsys := range f.f {
		if !ExistsFS(fsys, name) {
			continue
		}

		return fsys.Open(name)
	}

	return nil, fs.ErrNotExist
}

// SPDX-License-Identifier: MIT

package filesystem

import "io/fs"

type multipleFS struct {
	f []fs.FS
}

// MultipleFS 将多个 fs.FS 合并成一个 fs.FS 对象
//
// 每次查找时，都顺序查找第一个存在的对象并返回。
func MultipleFS(f ...fs.FS) fs.FS {
	return &multipleFS{f: f}
}

func (f *multipleFS) Open(name string) (fs.File, error) {
	for _, fsys := range f.f {
		if !ExistsFS(fsys, name) {
			continue
		}

		return fsys.Open(name)
	}

	return nil, fs.ErrNotExist
}

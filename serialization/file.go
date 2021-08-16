// SPDX-License-Identifier: MIT

package serialization

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Files 提供针对文件的序列化操作
type Files struct {
	*Serialization
}

// NewFiles 返回 Files 实例
func NewFiles(c int) *Files { return &Files{Serialization: New(c)} }

// Save 保存 v 到文件 p
func (f *Files) Save(p string, v interface{}) error {
	m, _ := f.searchByExt(p)
	if m == nil {
		return fmt.Errorf("未找到适合 %s 的函数", p)
	}

	data, err := m(v)
	if err != nil {
		return err
	}

	return os.WriteFile(p, data, fs.ModePerm)
}

// Load 加载文件到 v
func (f *Files) Load(p string, v interface{}) error {
	dir := filepath.ToSlash(filepath.Dir(p))
	name := filepath.ToSlash(filepath.Base(p))
	return f.LoadFS(os.DirFS(dir), name, v)
}

// LoadFS 加载文件到 v
func (f *Files) LoadFS(fsys fs.FS, name string, v interface{}) error {
	_, u := f.searchByExt(name)
	if u == nil {
		return fmt.Errorf("未找到适合 %s 的函数", name)
	}

	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		return err
	}

	return u(data, v)
}

func (f *Files) searchByExt(filename string) (MarshalFunc, UnmarshalFunc) {
	ext := filepath.Ext(filename)
	_, m, u := f.SearchFunc(func(s string) bool { return s == ext })
	return m, u
}

// SPDX-License-Identifier: MIT

package server

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/issue9/localeutil"
)

type MarshalFileFunc func(any) ([]byte, error)

type UnmarshalFileFunc func([]byte, any) error

type fileSerializer struct {
	m MarshalFileFunc
	u UnmarshalFileFunc
}

// Files 配置文件管理
type Files struct {
	server *Server
	items  map[string]*fileSerializer
}

func newFiles(s *Server) *Files {
	return &Files{
		server: s,
		items:  make(map[string]*fileSerializer, 5),
	}
}

// Files 配置文件操作接口
func (srv *Server) Files() *Files { return srv.files }

func (f *Files) Len() int { return len(f.items) }

// Add 添加新的序列方法
//
// ext 为文件扩展名，需要带 . 符号；
func (f *Files) Add(m MarshalFileFunc, u UnmarshalFileFunc, ext ...string) {
	if len(ext) == 0 {
		panic("参数 ext 不能为空")
	}

	for _, e := range ext {
		if _, found := f.items[e]; found {
			panic(fmt.Sprintf("已经存在同名的扩展名 %s", e))
		}
		f.items[e] = &fileSerializer{m: m, u: u}
	}
}

// Set 修改序列化方法
func (f *Files) Set(ext string, m MarshalFileFunc, u UnmarshalFileFunc) {
	f.items[ext] = &fileSerializer{m: m, u: u}
}

// Delete 删除序列化方法
func (f *Files) Delete(ext ...string) {
	for _, e := range ext {
		delete(f.items, e)
	}
}

// LoadLocale 加载本地化的文件
func (f *Files) LoadLocales(fsys fs.FS, glob string) error {
	matches, err := fs.Glob(fsys, glob)
	if err != nil {
		return err
	}

	if fsys == nil {
		fsys = f.server
	}

	for _, m := range matches {
		s := f.searchByExt(m)
		if s == nil {
			return localeutil.Error("not found serialization function for %s", m)
		}

		if err := localeutil.LoadMessageFromFS(f.server.CatalogBuilder(), fsys, m, s.u); err != nil {
			return err
		}
	}

	return nil
}

// Load 加载指定名称的文件内容至 v
//
// 根据文件扩展名决定采用什么编码方法；
func (f *Files) Load(fsys fs.FS, name string, v any) error {
	s := f.searchByExt(name)
	if s == nil {
		return localeutil.Error("not found serialization function for %s", name)
	}

	if fsys == nil {
		fsys = f.server
	}

	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		return err
	}

	return s.u(data, v)
}

// Save 将 v 解码并保存至 name 中
//
// 根据文件扩展名决定采用什么编码方法；
func (f *Files) Save(path string, v any) error {
	s := f.searchByExt(path)
	if s == nil {
		return localeutil.Error("not found serialization function for %s", path)
	}

	data, err := s.m(v)
	if err == nil {
		err = os.WriteFile(path, data, os.ModePerm)
	}
	return err
}

func (f *Files) searchByExt(name string) *fileSerializer {
	return f.items[filepath.Ext(name)]
}

// SPDX-License-Identifier: MIT

package server

import (
	"io/fs"
	"os"
	"path/filepath"
	"unicode"

	"github.com/issue9/cache"
	"github.com/issue9/sliceutil"
)

type Module struct {
	srv *Server
	id  string
	fs  []fs.FS
}

func isValidID(id string) bool {
	for _, b := range id {
		if unicode.IsSpace(b) || b == filepath.Separator || b == '/' {
			return false
		}
	}
	return fs.ValidPath(id) // 会根据 id 创建 fs.FS，所以必须符合 ValidPath
}

// NewModule 声明新的模块
//
// id 模块的 ID，需要全局唯一。会根据此值从 Server 派生出子文件系统。
func (srv *Server) NewModule(id string) *Module {
	if !isValidID(id) {
		panic("无效的 id 格式。")
	}

	contains := sliceutil.Index(srv.modules, func(i int) bool {
		return srv.modules[i] == id
	}) >= 0
	if contains {
		panic("存在同名模块")
	}

	f := make([]fs.FS, 0, 2)
	if fsys, err := fs.Sub(srv, id); err != nil {
		srv.Logs().Error(err) // 不退出
	} else {
		f = append(f, fsys)
	}

	srv.modules = append(srv.modules, id)
	return &Module{
		srv: srv,
		fs:  f,
		id:  id,
	}
}

// ID 模块的唯一 ID
func (m *Module) ID() string { return m.id }

func (m *Module) Server() *Server { return m.srv }

// AddFS 添加文件系统
//
// Module 默认以 id 为名称相对于 Server 创建了一个文件系统，
// 此操作会将 fsys 作为 Module 另一个文件系统与 Module 相关联，
// 当执行 Open 等操作时，会依然以关联顺序查找相应的文件系统，直到找到。
//
// 需要注意的是，fs.Glob 不是搜索所有的 fsys 然后返回集合。
func (m *Module) AddFS(fsys ...fs.FS) { m.fs = append(m.fs, fsys...) }

func (m *Module) Open(name string) (fs.File, error) {
	for _, fsys := range m.fs {
		if existsFS(fsys, name) {
			return fsys.Open(name)
		}
	}
	return nil, fs.ErrNotExist
}

func (m *Module) Glob(pattern string) ([]string, error) {
	for _, fsys := range m.fs {
		if matches, err := fs.Glob(fsys, pattern); len(matches) > 0 {
			return matches, err
		}
	}
	return nil, nil
}

// Cache 获取缓存对象
//
// 该缓存对象的 key 会自动添加 Module.ID 作为其前缀。
func (m *Module) Cache() cache.Access {
	return cache.Prefix(m.ID(), m.Server().Cache())
}

func existsFS(fsys fs.FS, p string) bool {
	_, err := fs.Stat(fsys, p)
	return err == nil || os.IsExist(err)
}

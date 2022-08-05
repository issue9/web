// SPDX-License-Identifier: MIT

package server

import (
	"io/fs"

	"github.com/issue9/cache"
	"github.com/issue9/sliceutil"
)

type Module struct {
	srv *Server
	id  string
	fs  []fs.FS
}

// IsValidID 是否为合法的 ID
//
// ID 只能是字母、数字、_ 以及 -
func IsValidID(id string) bool {
	if len(id) == 0 {
		return false
	}

	for _, c := range id {
		ok := c == '_' ||
			c == '-' ||
			(c >= '0' && c <= '9') ||
			(c >= 'a' && c <= 'z') ||
			(c >= 'A' && c <= 'Z')
		if !ok {
			return false
		}
	}
	return true
}

// NewModule 声明新的模块
//
// id 模块的 ID，需要全局唯一，只能是字母、数字以及下划线。
func (srv *Server) NewModule(id string) *Module {
	if !IsValidID(id) {
		panic("无效的 id 格式")
	}

	if sliceutil.Exists(srv.modules, func(e string) bool { return e == id }) {
		panic("存在同名模块：" + id)
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
		id:  id,
		fs:  f,
	}
}

// ID 模块的唯一 ID
func (m *Module) ID() string { return m.id }

// BuildID 根据当前模块的 ID 生成一个新的 ID
func (m *Module) BuildID(suffix string) string { return m.ID() + "_" + suffix }

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
		if f, err := fsys.Open(name); err == nil {
			return f, nil
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
// 该缓存对象的 key 会自动添加 [Module.ID] 作为其前缀。
func (m *Module) Cache() cache.Access {
	return cache.Prefix(m.BuildID(""), m.Server().Cache())
}

// LoadLocale 加载当前模块文件系统下的本地化文件
func (m *Module) LoadLocale(glob string) error {
	return m.Server().LoadLocale(m, glob)
}

// FileServer 返回以当前模块作为文件系统的静态文件服务
func (m *Module) FileServer(name, index string) HandlerFunc {
	return m.Server().FileServer(m, name, index)
}

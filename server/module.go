// SPDX-License-Identifier: MIT

package server

import (
	"io/fs"

	"github.com/issue9/sliceutil"

	"github.com/issue9/web/internal/filesystem"
)

// Module 相对独立的代码模块
//
// 模块带有一个唯一 ID 标记，所有通过模块向 Server 注册的内容，都会添加该 ID 值，
// 比如 Module.AddResult。
type Module struct {
	srv *Server
	id  string
	fs  *filesystem.MultipleFS
}

func (srv *Server) NewModule(id string) *Module {
	contains := sliceutil.Index(srv.modules, func(i int) bool {
		return srv.modules[i] == id
	}) >= 0
	if contains {
		panic("存在同名模块")
	}

	var f *filesystem.MultipleFS
	fsys, err := fs.Sub(srv, id)
	if err != nil {
		srv.Logs().Error(err) // 不退出，创建一个空的 filesystem.MultipleFS
		f = filesystem.NewMultipleFS()
	} else {
		f = filesystem.NewMultipleFS(fsys)
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
// 当执行 Open 等操作时，会依然以关联顺序查找相应的文件系统， 直到找到。
func (m *Module) AddFS(fsys ...fs.FS) { m.fs.Add(fsys...) }

func (m *Module) Open(name string) (fs.File, error) { return m.fs.Open(name) }

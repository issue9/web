// SPDX-License-Identifier: MIT

package module

import (
	"errors"
	"log"
	"strings"

	"github.com/issue9/scheduled"

	"github.com/issue9/web/context"
)

// ErrInited 当模块被多次初始化时返回此错误
var ErrInited = errors.New("模块已经初始化")

// Server 提供模块管理功能
type Server struct {
	ctxServer *context.Server

	// modules
	services  []*Service
	scheduled *scheduled.Server
	modules   []*Module
	inited    bool
}

// NewServer 声明一个新的 Server 实例
func NewServer(server *context.Server, plugin string) (*Server, error) {
	srv := &Server{
		ctxServer: server,

		services:  make([]*Service, 0, 100),
		scheduled: scheduled.NewServer(server.Location),
		modules:   make([]*Module, 0, 10),
	}

	srv.AddService(srv.scheduledService, "计划任务")

	if plugin != "" {
		if err := srv.loadPlugins(plugin); err != nil {
			return nil, err
		}
	}

	return srv, nil
}

// Init 初始化模块
//
// 若指定了 tag 参数，则只初始化与该标签相关联的内容。
//
// 一旦初始化完成，则不再接受添加新模块，也不能再次进行初始化。
// Server 和 Module 之中的大部分功能将失去操作意义，比如 Server.NewModule
// 虽然能添加新模块到 Server，但并不能真正初始化新的模块并挂载。
func (srv *Server) Init(tag string, info *log.Logger) error {
	if srv.inited && tag == "" {
		return ErrInited
	}

	flag := info.Flags()
	info.SetFlags(0)
	defer info.SetFlags(flag)

	info.Println("开始初始化模块...")

	if err := newDepencency(srv.modules, info).init(tag); err != nil {
		return err
	}

	if all := srv.ctxServer.Router().Mux().All(true, true); len(all) > 0 {
		info.Println("模块加载了以下路由项：")
		for path, methods := range all {
			info.Printf("[%s] %s\n", strings.Join(methods, ", "), path)
		}
	}

	info.Println("模块初始化完成！")

	if tag == "" {
		srv.inited = true
	}

	return nil
}

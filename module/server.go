// SPDX-License-Identifier: MIT

package module

import (
	"github.com/issue9/scheduled"

	"github.com/issue9/web/context"
)

// Server 提供模块管理功能
type Server struct {
	ctxServer *context.Server

	// modules
	services  []*Service
	scheduled *scheduled.Server
	modules   []*Module
}

// NewServer 声明一个新的 Modules 实例
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

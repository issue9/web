// SPDX-License-Identifier: MIT

package service

import (
	"context"

	"github.com/issue9/logs/v4"
	"github.com/issue9/scheduled"

	"github.com/issue9/web/internal/base"
)

type Server struct {
	services  []*Service
	scheduled *scheduled.Server
	base      *base.Base
	running   bool
}

// InternalNewServer 声明服务管理对象 Server
func InternalNewServer(loc *base.Base) *Server {
	return &Server{
		services:  make([]*Service, 0, 10),
		scheduled: scheduled.NewServer(loc.Location),
		base:      loc,
	}
}

func (srv *Server) Running() bool { return srv.running }

func (srv *Server) Run() {
	msg := srv.base.Printer.Sprintf("scheduled job")
	srv.Add(msg, func(ctx context.Context) error {
		go func() {
			l := srv.base.Logs
			if err := srv.scheduled.Serve(l.StdLogger(logs.LevelError), l.StdLogger(logs.LevelDebug)); err != nil {
				l.ERROR().Error(err)
			}
		}()

		<-ctx.Done()
		srv.scheduled.Stop()
		return context.Canceled
	})

	for _, s := range srv.services {
		s.Run()
	}

	srv.running = true
}

// Stop 停止所有服务
func (srv *Server) Stop() {
	for _, s := range srv.services {
		s.Stop()
	}
	srv.running = false
}

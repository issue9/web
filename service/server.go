// SPDX-License-Identifier: MIT

package service

import (
	"context"

	"github.com/issue9/logs/v4"
	"github.com/issue9/scheduled"

	"github.com/issue9/web/internal/locale"
)

type Server struct {
	services  []*Service
	scheduled *scheduled.Server
	logs      *logs.Logs
	locale    *locale.Locale
	running   bool
}

// InternalNewServer 声明服务管理对象 Server
func InternalNewServer(logs *logs.Logs, loc *locale.Locale) *Server {
	return &Server{
		services:  make([]*Service, 0, 10),
		scheduled: scheduled.NewServer(loc.Location),
		logs:      logs,
		locale:    loc,
	}
}

func (srv *Server) Running() bool { return srv.running }

func (srv *Server) Run() {
	l := srv.logs
	msg := srv.locale.Printer.Sprintf("scheduled job")
	srv.Add(msg, func(ctx context.Context) error {
		go func() {
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

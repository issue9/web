// SPDX-License-Identifier: MIT

package service

import (
	"context"
	"time"

	"github.com/issue9/logs/v4"
	"github.com/issue9/scheduled"
	"golang.org/x/text/message"
)

type Server struct {
	services  []*Service
	scheduled *scheduled.Server
	logs      *logs.Logs
	p         *message.Printer
	running   bool
}

func NewServer(logs *logs.Logs, loc *time.Location, p *message.Printer) *Server {
	return &Server{
		services:  make([]*Service, 0, 10),
		scheduled: scheduled.NewServer(loc),
		logs:      logs,
		p:         p,
	}
}

func (srv *Server) Run() {
	l := srv.logs
	msg := srv.p.Sprintf("scheduled job")
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

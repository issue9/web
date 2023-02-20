// SPDX-License-Identifier: MIT

package service

import (
	"context"
	"time"

	"github.com/issue9/localeutil"
	"github.com/issue9/scheduled"

	"github.com/issue9/web/logs"
)

type Server struct {
	ctx        context.Context
	cancelFunc context.CancelFunc

	services  []*Service
	scheduled *scheduled.Server
	running   bool
	errlog    logs.Logger
}

func NewServer(loc *time.Location, logs *logs.Logs) *Server {
	s := &Server{
		services:  make([]*Service, 0, 10),
		scheduled: scheduled.NewServer(loc, logs.ERROR().StdLogger(), logs.DEBUG().StdLogger()),
		errlog:    logs.ERROR(),
	}

	s.AddFunc(localeutil.Phrase("scheduled job"), s.scheduled.Serve)

	return s
}

func (srv *Server) Run() {
	srv.running = true

	// 在子项运行之前，重新生成 ctx
	srv.ctx, srv.cancelFunc = context.WithCancel(context.Background())
	for _, s := range srv.services {
		s.run()
	}
}

// Stop 停止所有服务
func (srv *Server) Stop() {
	srv.cancelFunc()
	srv.running = false
}

func (srv *Server) AddCron(title string, f JobFunc, spec string, delay bool) {
	srv.scheduled.Cron(title, f, spec, delay)
}

func (srv *Server) AddTicker(title string, f JobFunc, dur time.Duration, imm, delay bool) {
	srv.scheduled.Tick(title, f, dur, imm, delay)
}

func (srv *Server) AddAt(title string, f JobFunc, ti time.Time, delay bool) {
	srv.scheduled.At(title, f, ti, delay)
}

func (srv *Server) AddJob(title string, f JobFunc, scheduler Scheduler, delay bool) {
	srv.scheduled.New(title, f, scheduler, delay)
}

func (srv *Server) Jobs() []*Job { return srv.scheduled.Jobs() }

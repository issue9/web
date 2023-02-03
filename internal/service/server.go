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

func (srv *Server) Running() bool { return srv.running }

func (srv *Server) Run() {
	srv.running = true

	// 在子项运行之前，重新生成 ctx
	srv.ctx, srv.cancelFunc = context.WithCancel(context.Background())
	for _, s := range srv.services {
		s.Run()
	}
}

// Stop 停止所有服务
func (srv *Server) Stop() {
	srv.cancelFunc()
	srv.running = false
}

// AddCron 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// spec cron 表达式，支持秒；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (srv *Server) AddCron(title string, f scheduled.JobFunc, spec string, delay bool) {
	srv.scheduled.Cron(title, f, spec, delay)
}

// AddTicker 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// dur 时间间隔；
// imm 是否立即执行一次该任务；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (srv *Server) AddTicker(title string, f scheduled.JobFunc, dur time.Duration, imm, delay bool) {
	srv.scheduled.Tick(title, f, dur, imm, delay)
}

// AddAt 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// t 指定的时间点；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (srv *Server) AddAt(title string, f scheduled.JobFunc, ti time.Time, delay bool) {
	srv.scheduled.At(title, f, ti, delay)
}

// AddJob 添加新的计划任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// scheduler 计划任务的时间调度算法实现；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (srv *Server) AddJob(title string, f scheduled.JobFunc, scheduler scheduled.Scheduler, delay bool) {
	srv.scheduled.New(title, f, scheduler, delay)
}

// Jobs 返回所有的计划任务
func (srv *Server) Jobs() []*scheduled.Job { return srv.scheduled.Jobs() }

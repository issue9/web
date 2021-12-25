// SPDX-License-Identifier: MIT

package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/issue9/scheduled"
)

// 几种可能的状态值
const (
	ServiceStopped = scheduled.Stopped // 当前处于停止状态，默认状态
	ServiceRunning = scheduled.Running // 正在运行
	ServiceFailed  = scheduled.Failed  // 出错，不再执行后续操作
)

type (
	// ServiceFunc 服务实际需要执行的函数
	//
	// 实现者需要正确处理 ctx.Done 事件，调用者可能会主动取消函数执行；
	// 如果是通 ctx.Done 取消的，应该返回 context.Canceled。
	ServiceFunc func(ctx context.Context) error

	// Service 服务模型
	Service struct {
		srv        *Server
		Title      string
		f          ServiceFunc
		cancelFunc context.CancelFunc
		err        error // 保存上次的出错内容

		state    ServiceState
		stateMux sync.Mutex
	}

	// ServiceState 服务的状态值
	ServiceState = scheduled.State

	ScheduledJobFunc = scheduled.JobFunc
	ScheduledJob     = scheduled.Job
	Scheduler        = scheduled.Scheduler
)

// AddService 添加新的服务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明。
//
// NOTE: 如果 Manager 的所有服务已经处于运行的状态，则会自动运行新添加的服务。
func (srv *Server) AddService(title string, f ServiceFunc) {
	s := &Service{
		srv:   srv,
		Title: title,
		f:     f,
	}
	srv.services = append(srv.services, s)

	if srv.Serving() {
		s.Run()
	}
}

func (srv *Server) runServices() {
	l := srv.Logs()
	msg := srv.LocalePrinter().Sprintf("scheduled job")
	srv.AddService(msg, func(ctx context.Context) error {
		go func() {
			if err := srv.scheduled.Serve(l.ERROR(), l.DEBUG()); err != nil {
				l.Error(err)
			}
		}()

		<-ctx.Done()
		srv.scheduled.Stop()
		return context.Canceled
	})

	for _, s := range srv.services {
		s.Run()
	}
}

// Stop 停止所有服务
func (srv *Server) stopServices() {
	for _, s := range srv.services {
		s.Stop()
	}
}

// Services 返回长期运行的服务函数列表
func (srv *Server) Services() []*Service { return srv.services }

// State 获取当前服务的状态
func (srv *Service) State() ServiceState { return srv.state }

// Err 上次的错误信息，不会清空。
func (srv *Service) Err() error { return srv.err }

// Run 开始执行该服务
func (srv *Service) Run() {
	if srv.state == ServiceRunning {
		return
	}

	srv.stateMux.Lock()
	srv.state = ServiceRunning
	srv.stateMux.Unlock()

	go srv.serve()
}

func (srv *Service) serve() {
	defer func() {
		if msg := recover(); msg != nil {
			srv.err = fmt.Errorf("panic:%v", msg)
			srv.srv.logs.Error(srv.err)

			srv.stateMux.Lock()
			srv.state = ServiceFailed
			srv.stateMux.Unlock()
		}
	}()

	ctx := context.Background()
	ctx, srv.cancelFunc = context.WithCancel(ctx)
	srv.err = srv.f(ctx)
	state := ServiceStopped
	if srv.err != nil && srv.err != context.Canceled {
		srv.srv.logs.Error(srv.err)
		state = ServiceFailed
	}

	srv.stateMux.Lock()
	srv.state = state
	srv.stateMux.Unlock()
}

// Stop 停止服务
func (srv *Service) Stop() {
	if srv.state != ServiceRunning {
		return
	}

	if srv.cancelFunc != nil {
		srv.cancelFunc()
		srv.cancelFunc = nil
	}
}

// AddCron 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// spec cron 表达式，支持秒；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (srv *Server) AddCron(title string, f ScheduledJobFunc, spec string, delay bool) {
	srv.scheduled.Cron(title, f, spec, delay)
}

// AddTicker 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// dur 时间间隔；
// imm 是否立即执行一次该任务；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (srv *Server) AddTicker(title string, f ScheduledJobFunc, dur time.Duration, imm, delay bool) {
	srv.scheduled.Tick(title, f, dur, imm, delay)
}

// AddAt 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// t 指定的时间点；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (srv *Server) AddAt(title string, f ScheduledJobFunc, ti time.Time, delay bool) {
	srv.scheduled.At(title, f, ti, delay)
}

// AddJob 添加新的计划任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// scheduler 计划任务的时间调度算法实现；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (srv *Server) AddJob(title string, f ScheduledJobFunc, scheduler Scheduler, delay bool) {
	srv.scheduled.New(title, f, scheduler, delay)
}

// Jobs 返回所有的计划任务
func (srv *Server) Jobs() []*ScheduledJob { return srv.scheduled.Jobs() }

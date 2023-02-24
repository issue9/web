// SPDX-License-Identifier: MIT

package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/issue9/localeutil"
	"github.com/issue9/scheduled"
	"golang.org/x/text/message"
)

// 服务的几种状态
const (
	Stopped = scheduled.Stopped // 停止状态，默认状态
	Running = scheduled.Running // 正在运行
	Failed  = scheduled.Failed  // 出错，不再执行后续操作
)

type (
	Service struct {
		s       *Services
		title   localeutil.LocaleStringer
		service Servicer
		err     error // 保存上次的出错内容

		state    scheduled.State
		stateMux sync.Mutex
	}

	// Servicer 长期运行的服务需要实现的接口
	Servicer interface {
		// Serve 运行服务
		//
		// 这是个阻塞方法，实现者需要正确处理 [context.Context.Done] 事件。
		// 如果是通过 [context.Context] 的相关操作取消的，应该返回 [context.Context.Err]。
		Serve(context.Context) error
	}

	Services struct {
		s *Server
	}

	ServiceFunc func(context.Context) error

	State = scheduled.State

	Job = scheduled.Job

	JobFunc = scheduled.JobFunc

	Scheduler = scheduled.Scheduler
)

func (f ServiceFunc) Serve(ctx context.Context) error { return f(ctx) }

// Title 服务名称
func (srv *Service) Title(p *message.Printer) string {
	return srv.title.LocaleString(p)
}

// State 服务状态
func (srv *Service) State() State { return srv.state }

// Err 上次的错误信息
//
// 不会清空该值。
func (srv *Service) Err() error { return srv.err }

func (srv *Service) setState(s State) {
	srv.stateMux.Lock()
	srv.state = s
	srv.stateMux.Unlock()
}

func (srv *Service) run() {
	if srv.state != Running {
		srv.setState(Running)
		go srv.serve()
	}
}

func (srv *Service) serve() {
	defer func() {
		if msg := recover(); msg != nil {
			srv.err = fmt.Errorf("panic:%v", msg)
			srv.s.s.Logs().ERROR().Error(srv.err)
			srv.setState(Failed)
		}
	}()
	srv.err = srv.service.Serve(srv.s.s.ctx)
	state := Stopped
	if srv.err != nil && srv.err != context.Canceled {
		srv.s.s.Logs().ERROR().Error(srv.err)
		state = Failed
	}

	srv.setState(state)
}

// Services 服务管理
//
// 在 [Server] 初始之后，所有的服务就处于运行状态，后续添加的服务也会自动运行。
func (srv *Server) Services() *Services { return &Services{s: srv} }

// Add 添加并运行新的服务
//
// title 是对该服务的简要说明；
// f 表示服务的运行函数；
//
// NOTE: 如果服务已经处于运行的状态，则会自动运行新添加的服务。
func (srv *Services) Add(title localeutil.LocaleStringer, f Servicer) {
	s := &Service{
		s:       srv,
		title:   title,
		service: f,
	}
	srv.s.services = append(srv.s.services, s)
	s.run()
}

func (srv *Services) AddFunc(title localeutil.LocaleStringer, f func(context.Context) error) {
	srv.Add(title, ServiceFunc(f))
}

func (srv *Services) Services() []*Service { return srv.s.services }

// AddCron 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// spec cron 表达式，支持秒；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (srv *Services) AddCron(title string, f JobFunc, spec string, delay bool) {
	srv.s.scheduled.Cron(title, f, spec, delay)
}

// AddTicker 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// dur 时间间隔；
// imm 是否立即执行一次该任务；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (srv *Services) AddTicker(title string, f JobFunc, dur time.Duration, imm, delay bool) {
	srv.s.scheduled.Tick(title, f, dur, imm, delay)
}

// AddAt 添加新的定时任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// t 指定的时间点；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (srv *Services) AddAt(title string, f JobFunc, ti time.Time, delay bool) {
	srv.s.scheduled.At(title, f, ti, delay)
}

// AddJob 添加新的计划任务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明；
// scheduler 计划任务的时间调度算法实现；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (srv *Services) AddJob(title string, f JobFunc, scheduler Scheduler, delay bool) {
	srv.s.scheduled.New(title, f, scheduler, delay)
}

// Jobs 返回所有的计划任务
func (srv *Services) Jobs() []*Job { return srv.s.scheduled.Jobs() }

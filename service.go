// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/issue9/scheduled"
)

// 服务的几种状态
const (
	Stopped State = scheduled.Stopped // 停止状态，默认状态
	Running State = scheduled.Running // 正在运行
	Failed  State = scheduled.Failed  // 出错，不再执行后续操作
)

type (
	// Services [Service] 管理接口
	Services struct {
		server    *Server
		services  []*Service
		scheduled *scheduled.Server
		wait      *sync.WaitGroup
	}

	// Service 单个服务对象
	Service struct {
		s       *Services
		title   LocaleStringer
		service ServiceFunc
		err     error // 保存上次的出错内容，不会清空该值。

		state    State
		stateMux sync.RWMutex
	}

	// ServiceFunc 服务函数
	//
	// 这是个阻塞方法，实现者需要正确处理 [context.Context.Done] 事件。
	// 如果是通过 [context.Context] 的相关操作取消的，应该返回 [context.Context.Err]。
	ServiceFunc func(context.Context) error

	// State 服务状态
	State = scheduled.State

	Job           = scheduled.Job
	JobFunc       = scheduled.JobFunc
	Scheduler     = scheduled.Scheduler
	SchedulerFunc = scheduled.SchedulerFunc
)

// Services 服务管理接口
func (s *Server) Services() *Services { return s.services }

func (s *Server) initServices() {
	s.services = &Services{
		server:    s,
		services:  make([]*Service, 0, 5),
		scheduled: scheduled.NewServer(s.Location(), s.Logs().ERROR(), s.Logs().DEBUG()),
		wait:      &sync.WaitGroup{},
	}

	f := func(ctx context.Context) error { return s.services.scheduled.Serve(ctx) }
	s.Services().Add(StringPhrase("scheduler jobs"), f)
}

// Title 服务的描述信息
func (srv *Service) Title() LocaleStringer { return srv.title }

// Err 服务的出错时的错误信息
func (srv *Service) Err() error { return srv.err }

// State 服务的当前状态
func (srv *Service) State() State {
	srv.stateMux.RLock()
	s := srv.state
	srv.stateMux.RUnlock()
	return s
}

func (srv *Service) setState(s State) {
	srv.stateMux.Lock()
	srv.state = s
	srv.stateMux.Unlock()
}

func (srv *Service) goServe(ctx context.Context) {
	if srv.State() != Running {
		srv.setState(Running)
		srv.s.wait.Go(func() { srv.serve(ctx) })
	}
}

func (srv *Service) serve(ctx context.Context) {
	defer func() {
		if msg := recover(); msg != nil {
			srv.err = fmt.Errorf("panic:%v", msg)
			srv.s.server.Logs().ERROR().Error(srv.err)
			srv.setState(Failed)
		}
	}()
	srv.err = srv.service(ctx)
	state := Stopped
	if !errors.Is(srv.err, context.Canceled) && !errors.Is(srv.err, http.ErrServerClosed) {
		srv.s.server.Logs().ERROR().Error(srv.err)
		state = Failed
	}

	srv.setState(state)
}

// Add 添加并运行新的服务
//
// title 是对该服务的简要说明；
// 返回取消该服务的操作函数，该函数同时会将整个服务从列表中删除；
func (srv *Services) Add(title LocaleStringer, f ServiceFunc) context.CancelFunc {
	s := &Service{
		s:       srv,
		title:   title,
		service: f,
	}
	srv.services = append(srv.services, s)

	ctx, cancel := context.WithCancel(srv.server)
	s.goServe(ctx)

	return func() {
		s.setState(Stopped)
		cancel()
		srv.services = slices.DeleteFunc(srv.services, func(e *Service) bool { return e == s })
	}
}

// AddFunc 将函数 f 作为服务添加并运行
func (srv *Services) AddFunc(title LocaleStringer, f func(context.Context) error) context.CancelFunc {
	return srv.Add(title, ServiceFunc(f))
}

// AddCron 添加新的定时任务
//
// title 是对该服务的简要说明；
// spec cron 表达式，支持秒；
func (srv *Services) AddCron(title LocaleStringer, f JobFunc, spec string) context.CancelFunc {
	return srv.scheduled.Cron(title, f, spec)
}

// AddTicker 添加新的定时任务
//
// title 是对该服务的简要说明；
// dur 时间间隔；
// imm 是否先执行一次该任务，如果为 true 将会排在任务执行任务队列的前列，而不是立即执行；
func (srv *Services) AddTicker(title LocaleStringer, job JobFunc, dur time.Duration, imm bool) context.CancelFunc {
	return srv.scheduled.Tick(title, job, dur, imm)
}

// AddAt 添加在某个时间点执行的任务
//
// title 是对该服务的简要说明；
// at 指定的时间点；
func (srv *Services) AddAt(title LocaleStringer, job JobFunc, at time.Time) context.CancelFunc {
	return srv.scheduled.At(title, job, at)
}

// AddJob 添加新的计划任务
//
// title 是对该服务的简要说明；
// scheduler 计划任务的时间调度算法实现；
func (srv *Services) AddJob(title LocaleStringer, job JobFunc, scheduler Scheduler) context.CancelFunc {
	return srv.scheduled.New(title, job, scheduler)
}

// Jobs 访问所有的计划任务
func (srv *Services) Jobs() iter.Seq[*scheduled.Job] { return srv.scheduled.Jobs() }

// All 返回所有的服务对象
func (srv *Services) All() iter.Seq[*Service] {
	return func(yield func(*Service) bool) {
		for _, s := range srv.services {
			if !yield(s) {
				break
			}
		}
	}
}

func (srv *Services) waitExit() { srv.wait.Wait() }

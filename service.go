// SPDX-License-Identifier: MIT

package web

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/issue9/scheduled"
)

// 服务的几种状态
const (
	Stopped = scheduled.Stopped // 停止状态，默认状态
	Running = scheduled.Running // 正在运行
	Failed  = scheduled.Failed  // 出错，不再执行后续操作
)

type (
	Services struct {
		s         *InternalServer
		services  []*service
		ctx       context.Context
		scheduled *scheduled.Server
	}

	service struct {
		s       *Services
		title   LocaleStringer
		service Service
		err     error // 保存上次的出错内容，不会清空该值。

		state    State
		stateMux sync.RWMutex
	}

	// Service 长期运行的服务需要实现的接口
	Service interface {
		// Serve 运行服务
		//
		// 这是个阻塞方法，实现者需要正确处理 [context.Context.Done] 事件。
		// 如果是通过 [context.Context] 的相关操作取消的，应该返回 [context.Context.Err]。
		Serve(context.Context) error
	}

	ServiceFunc func(context.Context) error

	// State 服务状态
	//
	// 以下设置用于 restdoc
	//
	// @type string
	// @enum stopped running failed
	State = scheduled.State

	Job           = scheduled.Job
	JobFunc       = scheduled.JobFunc
	Scheduler     = scheduled.Scheduler
	SchedulerFunc = scheduled.SchedulerFunc
)

func (f ServiceFunc) Serve(ctx context.Context) error { return f(ctx) }

func (s *InternalServer) Services() *Services { return s.services }

func (s *InternalServer) initServices() {
	ctx, cancel := context.WithCancelCause(context.Background())
	s.OnClose(func() error { cancel(http.ErrServerClosed); return nil })

	s.services = &Services{
		s:         s,
		services:  make([]*service, 0, 5),
		ctx:       ctx,
		scheduled: scheduled.NewServer(s.Location(), s.Logs().ERROR(), s.Logs().DEBUG()),
	}
	s.Services().Add(StringPhrase("scheduler jobs"), s.services.scheduled)
}

func (srv *service) getState() State {
	srv.stateMux.RLock()
	s := srv.state
	srv.stateMux.RUnlock()
	return s
}

func (srv *service) setState(s State) {
	srv.stateMux.Lock()
	srv.state = s
	srv.stateMux.Unlock()
}

func (srv *service) goServe(ctx context.Context) {
	if srv.getState() != Running {
		srv.setState(Running)
		go srv.serve(ctx)
	}
}

func (srv *service) serve(ctx context.Context) {
	defer func() {
		if msg := recover(); msg != nil {
			srv.err = fmt.Errorf("panic:%v", msg)
			srv.s.s.server.Logs().ERROR().Error(srv.err)
			srv.setState(Failed)
		}
	}()
	srv.err = srv.service.Serve(ctx)
	state := Stopped
	if !errors.Is(srv.err, context.Canceled) {
		srv.s.s.Logs().ERROR().Error(srv.err)
		state = Failed
	}

	srv.setState(state)
}

// Add 添加并运行新的服务
//
// title 是对该服务的简要说明；
// 返回取消该服务的操作函数，该函数同时会将整个服务从列表中删除；
func (srv *Services) Add(title LocaleStringer, f Service) context.CancelFunc {
	s := &service{
		s:       srv,
		title:   title,
		service: f,
	}
	srv.services = append(srv.services, s)

	ctx, c := context.WithCancel(srv.ctx)
	s.goServe(ctx)

	return func() {
		s.setState(Stopped)
		c()
		srv.services = slices.DeleteFunc(srv.services, func(e *service) bool { return e == s })
	}
}

// AddFunc 将函数 f 作为服务添加并运行
func (srv *Services) AddFunc(title LocaleStringer, f func(context.Context) error) context.CancelFunc {
	return srv.Add(title, ServiceFunc(f))
}

// Visit 访问所有的服务
//
// visit 的原型为：
//
//	func(title LocaleStringer, state State, err error)
//
// title 为服务的说明；
// state 为服务的当前状态；
// err 只在 state 为 [Failed] 时才有的错误说明；
func (srv *Services) Visit(visit func(title LocaleStringer, state State, err error)) {
	for _, s := range srv.services {
		ss := s // TODO(go1.22): 可省略
		visit(ss.title, ss.getState(), ss.err)
	}
}

// AddCron 添加新的定时任务
//
// title 是对该服务的简要说明；
// spec cron 表达式，支持秒；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (srv *Services) AddCron(title LocaleStringer, f JobFunc, spec string, delay bool) func() {
	return srv.scheduled.Cron(title, f, spec, delay)
}

// AddTicker 添加新的定时任务
//
// title 是对该服务的简要说明；
// dur 时间间隔；
// imm 是否立即执行一次该任务；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (srv *Services) AddTicker(title LocaleStringer, job JobFunc, dur time.Duration, imm, delay bool) func() {
	return srv.scheduled.Tick(title, job, dur, imm, delay)
}

// AddAt 添加在某个时间点执行的任务
//
// title 是对该服务的简要说明；
// at 指定的时间点；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (srv *Services) AddAt(title LocaleStringer, job JobFunc, at time.Time, delay bool) func() {
	return srv.scheduled.At(title, job, at, delay)
}

// AddJob 添加新的计划任务
//
// title 是对该服务的简要说明；
// scheduler 计划任务的时间调度算法实现；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
func (srv *Services) AddJob(title LocaleStringer, job JobFunc, scheduler Scheduler, delay bool) func() {
	return srv.scheduled.New(title, job, scheduler, delay)
}

// VisitJobs 访问所有的计划任务
func (srv *Services) VisitJobs(visit func(*Job)) {
	for _, j := range srv.scheduled.Jobs() {
		jj := j // TODO(go1.22): 可省略
		visit(jj)
	}
}

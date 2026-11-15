// SPDX-FileCopyrightText: 2018-2026 caixw
//
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
	Stopped   State = scheduled.Stopped   // 停止状态，默认状态
	Running   State = scheduled.Running   // 正在运行
	Failed    State = scheduled.Failed    // 出错，不再执行后续操作
	Cancelled State = scheduled.Cancelled // 由用户取消
)

type (
	// Services [Service] 管理接口
	Services struct {
		server    *Server
		scheduled *scheduled.Server
		wait      *sync.WaitGroup

		mux      sync.RWMutex
		services []*service
	}

	service struct {
		s       *Services
		title   LocaleStringer
		service ServiceFunc

		// 以下内容需要上锁

		mux   sync.RWMutex
		err   error // 保存上次的出错内容，不会清空该值。
		state State
	}

	// Service 单个服务对象的快照信息
	Service struct {
		Err   error
		Title LocaleStringer
		State State
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
		services:  make([]*service, 0, 5),
		scheduled: scheduled.NewServer(s.Location(), s.Logs().ERROR(), s.Logs().INFO()),
		wait:      &sync.WaitGroup{},
	}

	f := func(ctx context.Context) error { return s.services.scheduled.Serve(ctx) }
	s.Services().Add(StringPhrase("scheduler jobs"), f)
}

func (srv *service) setState(s State) {
	srv.mux.Lock()
	srv.state = s
	srv.mux.Unlock()
}

func (srv *service) setError(err error) {
	srv.mux.Lock()
	srv.err = err
	srv.mux.Unlock()
}

func (srv *service) goServe(ctx context.Context) {
	srv.mux.Lock()
	defer srv.mux.Unlock()

	if srv.state != Running {
		srv.state = Running
		srv.s.wait.Go(func() { srv.serve(ctx) })
	}
}

func (srv *service) serve(ctx context.Context) {
	defer func() {
		if msg := recover(); msg != nil {
			err := fmt.Errorf("panic:%v", msg)
			srv.s.server.Logs().ERROR().Error(err)

			srv.setError(err)
			srv.setState(Failed)
		}
	}()

	err := srv.service(ctx)
	state := Stopped
	if err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, http.ErrServerClosed) {
		srv.s.server.Logs().ERROR().Error(err)
		state = Failed
	}

	srv.setError(err)
	srv.setState(state)
}

// Add 添加并运行新的服务
//
// title 是对该服务的简要说明；
// 返回取消该服务的操作函数，该函数同时会将整个服务从列表中删除；
func (srv *Services) Add(title LocaleStringer, f ServiceFunc) context.CancelFunc {
	if srv.server.State() == Cancelled {
		panic("服务已经停止，不能再添加新服务")
	}

	s := &service{
		s:       srv,
		title:   title,
		service: f,
	}

	srv.mux.Lock()
	srv.services = append(srv.services, s)
	srv.mux.Unlock()

	ctx, cancel := context.WithCancel(srv.server)
	s.goServe(ctx)

	return func() {
		s.setState(Stopped)
		cancel()
		srv.mux.Lock()
		srv.services = slices.DeleteFunc(srv.services, func(e *service) bool { return e == s })
		srv.mux.Unlock()
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
// 具体说明可参考 [scheduled.Server.Tick]
func (srv *Services) AddTicker(title LocaleStringer, job JobFunc, dur time.Duration, imm bool) context.CancelFunc {
	return srv.scheduled.Tick(title, job, dur, imm)
}

// AddAt 添加在某个时间点执行的任务
//
// 具体说明可参考 [scheduled.Server.At]
func (srv *Services) AddAt(title LocaleStringer, job JobFunc, at time.Time) context.CancelFunc {
	return srv.scheduled.At(title, job, at)
}

// AddEarliest 添加 Earliest 计划任务
//
// 具体说明可参考 [scheduled.Server.Earliest]
func (srv *Services) AddEarliest(title LocaleStringer, job JobFunc, s ...scheduled.Scheduler) context.CancelFunc {
	return srv.scheduled.Earliest(title, job, s...)
}

// AddJob 添加新的计划任务
//
// 具体说明可参考 [scheduled.Server.New]
func (srv *Services) AddJob(title LocaleStringer, job JobFunc, scheduler Scheduler, imm bool) context.CancelFunc {
	return srv.scheduled.New(title, job, scheduler, imm)
}

// Jobs 访问所有的计划任务
func (srv *Services) Jobs() []*scheduled.Job { return srv.scheduled.Jobs() }

// All 返回所有服务对象的当前快照
func (srv *Services) All() []*Service {
	srv.mux.RLock()
	defer srv.mux.RUnlock()

	ss := make([]*Service, 0, len(srv.services))
	for _, s := range srv.services {
		s.mux.RLock()
		ss = append(ss, &Service{
			Err:   s.err,
			State: s.state,
			Title: s.title,
		})
		s.mux.RUnlock()
	}

	return ss
}

func (srv *Services) waitExit() { srv.wait.Wait() }

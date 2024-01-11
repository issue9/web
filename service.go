// SPDX-License-Identifier: MIT

package web

import (
	"context"
	"errors"
	"fmt"
	"net/http"
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
		jobTitles map[string]LocaleStringer
	}

	service struct {
		s       *Services
		title   LocaleStringer
		service Service
		err     error // 保存上次的出错内容，不会清空该值。

		state    State
		stateMux sync.Mutex
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
		jobTitles: make(map[string]LocaleStringer, 10),
	}
	s.services.Add(StringPhrase("scheduler jobs"), s.services.scheduled)
}

func (srv *service) setState(s State) {
	srv.stateMux.Lock()
	srv.state = s
	srv.stateMux.Unlock()
}

func (srv *service) goServe() {
	if srv.state != Running {
		srv.setState(Running)
		go srv.serve()
	}
}

func (srv *service) serve() {
	defer func() {
		if msg := recover(); msg != nil {
			srv.err = fmt.Errorf("panic:%v", msg)
			srv.s.s.server.Logs().ERROR().Error(srv.err)
			srv.setState(Failed)
		}
	}()
	srv.err = srv.service.Serve(srv.s.ctx)
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
func (srv *Services) Add(title LocaleStringer, f Service) {
	s := &service{
		s:       srv,
		title:   title,
		service: f,
	}
	srv.services = append(srv.services, s)
	s.goServe()
}

// AddFunc 将函数 f 作为服务添加并运行
func (srv *Services) AddFunc(title LocaleStringer, f func(context.Context) error) {
	srv.Add(title, ServiceFunc(f))
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
		visit(s.title, s.state, s.err)
	}
}

// AddCron 添加新的定时任务
//
// title 是对该服务的简要说明；
// spec cron 表达式，支持秒；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
//
// NOTE: 此功能依赖 [Server.UniqueID]。
func (srv *Services) AddCron(title LocaleStringer, f JobFunc, spec string, delay bool) {
	id := srv.s.UniqueID()
	srv.jobTitles[id] = title
	srv.scheduled.Cron(id, f, spec, delay)
}

// AddTicker 添加新的定时任务
//
// title 是对该服务的简要说明；
// dur 时间间隔；
// imm 是否立即执行一次该任务；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
//
// NOTE: 此功能依赖 [Server.UniqueID]。
func (srv *Services) AddTicker(title LocaleStringer, job JobFunc, dur time.Duration, imm, delay bool) {
	id := srv.s.UniqueID()
	srv.jobTitles[id] = title
	srv.scheduled.Tick(id, job, dur, imm, delay)
}

// AddAt 添加在某个时间点执行的任务
//
// title 是对该服务的简要说明；
// at 指定的时间点；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
//
// NOTE: 此功能依赖 [Server.UniqueID]。
func (srv *Services) AddAt(title LocaleStringer, job JobFunc, at time.Time, delay bool) {
	id := srv.s.UniqueID()
	srv.jobTitles[id] = title
	srv.scheduled.At(id, job, at, delay)
}

// AddJob 添加新的计划任务
//
// title 是对该服务的简要说明；
// scheduler 计划任务的时间调度算法实现；
// delay 是否在任务执行完之后，才计算下一次的执行时间点。
//
// NOTE: 此功能依赖 [Server.UniqueID]。
func (srv *Services) AddJob(title LocaleStringer, job JobFunc, scheduler Scheduler, delay bool) {
	id := srv.s.UniqueID()
	srv.jobTitles[id] = title
	srv.scheduled.New(id, job, scheduler, delay)
}

// VisitJobs 访问所有的计划任务
//
// visit 原型为：
//
//	func(title LocaleStringer, prev, next time.Time, state State, delay bool, err error)
//
// title 为计划任务的说明；
// prev 和 next 表示任务的上一次执行时间和下一次执行时间；
// state 表示当前的状态；
// delay 表示该任务是否是执行完才开始计算下一次任务时间的；
// err 表示这个任务的出错状态；
func (srv *Services) VisitJobs(visit func(LocaleStringer, time.Time, time.Time, State, bool, error)) {
	for _, j := range srv.scheduled.Jobs() {
		visit(srv.jobTitles[j.Name()], j.Prev(), j.Next(), j.State(), j.Delay(), j.Err())
	}
}

// SPDX-License-Identifier: MIT

// Package service 服务管理
package service

import (
	"context"
	"fmt"
	"sync"

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
		s       *Server
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

	Func func(context.Context) error

	State = scheduled.State

	Job = scheduled.Job

	JobFunc = scheduled.JobFunc

	Scheduler = scheduled.Scheduler
)

func (f Func) Serve(ctx context.Context) error { return f(ctx) }

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
			srv.s.errlog.Print(msg)
			srv.setState(Failed)
		}
	}()
	srv.err = srv.service.Serve(srv.s.ctx)
	state := Stopped
	if srv.err != nil && srv.err != context.Canceled {
		srv.s.errlog.Error(srv.err)
		state = Failed
	}

	srv.setState(state)
}

func (srv *Server) Add(title localeutil.LocaleStringer, f Servicer) {
	s := &Service{
		s:       srv,
		title:   title,
		service: f,
	}
	srv.services = append(srv.services, s)

	if srv.running {
		s.run()
	}
}

func (srv *Server) AddFunc(title localeutil.LocaleStringer, f func(context.Context) error) {
	srv.Add(title, Func(f))
}

func (srv *Server) Services() []*Service { return srv.services }

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

// ServiceFunc 服务函数的原型
//
// 实现者需要正确处理 ctx.Done 事件，调用者可能会主动取消函数执行；
// 如果是通 ctx.Done 取消的，应该返回 [context.Canceled]。
type ServiceFunc func(ctx context.Context) error

type (
	Service struct {
		s          *Server
		title      localeutil.LocaleStringer
		f          ServiceFunc
		cancelFunc context.CancelFunc
		err        error // 保存上次的出错内容

		state    scheduled.State
		stateMux sync.Mutex
	}
)

// Title 服务名称
func (srv *Service) Title(p *message.Printer) string {
	return srv.title.LocaleString(p)
}

// State 服务状态
func (srv *Service) State() scheduled.State { return srv.state }

// Err 上次的错误信息
//
// 不会清空该值。
func (srv *Service) Err() error { return srv.err }

func (srv *Service) setState(s scheduled.State) {
	srv.stateMux.Lock()
	srv.state = s
	srv.stateMux.Unlock()
}

// Run 开始执行该服务
func (srv *Service) Run() {
	if srv.state != scheduled.Running {
		srv.setState(scheduled.Running)
		go srv.serve()
	}
}

func (srv *Service) serve() {
	defer func() {
		if msg := recover(); msg != nil {
			srv.err = fmt.Errorf("panic:%v", msg)
			srv.s.logs.ERROR().Error(srv.err)
			srv.setState(scheduled.Failed)
		}
	}()

	ctx := context.Background()
	ctx, srv.cancelFunc = context.WithCancel(ctx)
	srv.err = srv.f(ctx)
	state := scheduled.Stopped
	if srv.err != nil && srv.err != context.Canceled {
		srv.s.logs.ERROR().Error(srv.err)
		state = scheduled.Failed
	}

	srv.setState(state)
}

func (srv *Service) Stop() {
	if srv.state != scheduled.Running {
		return
	}

	if srv.cancelFunc != nil {
		srv.cancelFunc()
		srv.cancelFunc = nil
	}
}

// Add 添加新的服务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明。
//
// NOTE: 如果所有服务已经处于运行的状态，则会自动运行新添加的服务。
func (srv *Server) Add(title localeutil.LocaleStringer, f ServiceFunc) {
	s := &Service{
		s:     srv,
		title: title,
		f:     f,
	}
	srv.services = append(srv.services, s)

	if srv.running {
		s.Run()
	}
}

func (srv *Server) Services() []*Service { return srv.services }

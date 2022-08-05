// SPDX-License-Identifier: MIT

// Package service 服务管理
package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/issue9/scheduled"
)

// 几种可能的状态值
const (
	Stopped = scheduled.Stopped // 当前处于停止状态，默认状态
	Running = scheduled.Running // 正在运行
	Failed  = scheduled.Failed  // 出错，不再执行后续操作
)

type (
	// ServiceFunc 服务实际需要执行的函数
	//
	// 实现者需要正确处理 ctx.Done 事件，调用者可能会主动取消函数执行；
	// 如果是通 ctx.Done 取消的，应该返回 [context.Canceled]。
	ServiceFunc func(ctx context.Context) error

	// Service 服务模型
	Service struct {
		srv        *Server
		title      string
		f          ServiceFunc
		cancelFunc context.CancelFunc
		err        error // 保存上次的出错内容

		state    State
		stateMux sync.Mutex
	}

	// State 服务的状态值
	State = scheduled.State
)

// Title 服务名称
func (srv *Service) Title() string { return srv.title }

// State 服务状态
func (srv *Service) State() State { return srv.state }

// Err 上次的错误信息
//
// 不会清空该值。
func (srv *Service) Err() error { return srv.err }

// Run 开始执行该服务
func (srv *Service) Run() {
	if srv.state == Running {
		return
	}

	srv.stateMux.Lock()
	srv.state = Running
	srv.stateMux.Unlock()

	go srv.serve()
}

func (srv *Service) serve() {
	defer func() {
		if msg := recover(); msg != nil {
			srv.err = fmt.Errorf("panic:%v", msg)
			srv.srv.logs.Error(srv.err)

			srv.stateMux.Lock()
			srv.state = Failed
			srv.stateMux.Unlock()
		}
	}()

	ctx := context.Background()
	ctx, srv.cancelFunc = context.WithCancel(ctx)
	srv.err = srv.f(ctx)
	state := Stopped
	if srv.err != nil && srv.err != context.Canceled {
		srv.srv.logs.Error(srv.err)
		state = Failed
	}

	srv.stateMux.Lock()
	srv.state = state
	srv.stateMux.Unlock()
}

// Stop 停止服务
func (srv *Service) Stop() {
	if srv.state != Running {
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
func (srv *Server) Add(title string, f ServiceFunc) {
	s := &Service{
		srv:   srv,
		title: title,
		f:     f,
	}
	srv.services = append(srv.services, s)

	if srv.running {
		s.Run()
	}
}

// Services 返回长期运行的服务函数列表
func (srv *Server) Services() []*Service { return srv.services }

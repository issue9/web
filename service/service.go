// SPDX-License-Identifier: MIT

// Package service 服务管理
package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/issue9/scheduled"
)

// Func 服务实际需要执行的函数
//
// 实现者需要正确处理 ctx.Done 事件，调用者可能会主动取消函数执行；
// 如果是通 ctx.Done 取消的，应该返回 context.Canceled。
type Func func(ctx context.Context) error

// State 服务的状态值
type State = scheduled.State

// 几种可能的状态值
const (
	Stopped = scheduled.Stopped // 当前处于停止状态，默认状态
	Running = scheduled.Running // 正在运行
	Failed  = scheduled.Failed  // 出错，不再执行后续操作
)

// Service 服务模型
type Service struct {
	mgr        *Manager
	Title      string
	f          Func
	cancelFunc context.CancelFunc
	err        error // 保存上次的出错内容

	state    State
	stateMux sync.Mutex
}

// AddService 添加新的服务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明。
//
// NOTE: 如果 Manager 的所有服务已经处于运行的状态，则会自动运行新添加的服务。
func (mgr *Manager) AddService(title string, f Func) {
	srv := &Service{
		mgr:   mgr,
		Title: title,
		f:     f,
	}
	mgr.services = append(mgr.services, srv)

	if mgr.running {
		srv.Run()
	}
}

// Services 返回所有的服务列表
func (mgr *Manager) Services() []*Service { return mgr.services }

// State 获取当前服务的状态
func (srv *Service) State() State { return srv.state }

// Err 上次的错误信息，不会清空。
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
			srv.mgr.logs.Error(srv.err)

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
		srv.mgr.logs.Error(srv.err)
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

// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package module

import (
	"context"
	"fmt"
	"sync"
)

// ServiceFunc 服务实际需要执行的函数
//
// 实现者需要正确处理 ctx.Done 事件，调用者可能会主动取消函数执行；
// 如果是通 ctx 取消的，应该返回其错误信息。
type ServiceFunc func(ctx context.Context) error

// ServiceState 服务的状态值
type ServiceState int8

// 几种可能的状态值
const (
	ServiceStop    ServiceState = iota + 1 // 当前处理停止状态，默认状态
	ServiceRunning                         // 正在运行
	ServiceFailed                          // 出错，不再执行后续操作
)

// Service 服务模型
type Service struct {
	Title  string
	Module *Module

	state      ServiceState
	f          ServiceFunc
	cancelFunc context.CancelFunc
	locker     sync.Mutex
	err        error // 保存上次的出错内容
}

// AddService 添加新的服务
//
// f 表示服务的运行函数；
// title 是对该服务的简要说明。
func (m *Module) AddService(f ServiceFunc, title string) {
	m.AddInit(func() error {
		m.ms.services = append(m.ms.services, &Service{
			Title:  title,
			Module: m,
			state:  ServiceStop,
			f:      f,
		})
		return nil
	}, "注册服务："+title)
}

// State 获取当前服务的状态
func (srv *Service) State() ServiceState {
	return srv.state
}

// Err 上次的错误信息，不会清空。
func (srv *Service) Err() error {
	return srv.err
}

// Run 开始执行该服务
func (srv *Service) Run() {
	if srv.state == ServiceRunning {
		return
	}

	srv.locker.Lock()
	defer srv.locker.Unlock()

	srv.state = ServiceRunning
	go srv.serve()
}

func (srv *Service) serve() {
	defer func() {
		if msg := recover(); msg != nil {
			srv.err = fmt.Errorf("panic:%v", msg)
			srv.locker.Lock()
			srv.state = ServiceFailed
			srv.locker.Unlock()
		}
	}()

	ctx := context.Background()
	ctx, srv.cancelFunc = context.WithCancel(ctx)
	srv.err = srv.f(ctx)
	state := ServiceStop
	if srv.err != nil && srv.err != context.Canceled {
		state = ServiceFailed
	}

	srv.locker.Lock()
	srv.state = state
	srv.locker.Unlock()
}

// Stop 停止服务。
func (srv *Service) Stop() {
	if srv.state != ServiceRunning {
		return
	}

	if srv.cancelFunc != nil {
		srv.cancelFunc()
		srv.cancelFunc = nil
	}
}

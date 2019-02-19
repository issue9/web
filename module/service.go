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
	ID    int // 唯一标志，在运行之后才会赋值
	Title string

	state      ServiceState
	f          ServiceFunc
	cancelFunc context.CancelFunc
	locker     sync.Mutex

	err error // 保存上次的出错内容
}

// AddService 添加新的服务
func (m *Module) AddService(f ServiceFunc, title string) {
	if m.services == nil {
		m.services = make([]*Service, 0, 5)
	}

	m.services = append(m.services, &Service{
		Title: title,
		state: ServiceStop,
		f:     f,
	})
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
	srv.locker.Lock()
	defer srv.locker.Unlock()

	if srv.state != ServiceRunning {
		go srv.serve()
	}
}

func (srv *Service) serve() {
	defer func() {
		if msg := recover(); msg != nil {
			srv.err = fmt.Errorf("panic:%v", msg)
			srv.state = ServiceFailed
		}
	}()

	ctx := context.Background()
	ctx, srv.cancelFunc = context.WithCancel(ctx)
	srv.state = ServiceRunning

	err := srv.f(ctx)
	if err != nil && err != context.Canceled {
		srv.err = err
		srv.state = ServiceFailed
		return
	}

	srv.state = ServiceStop
}

// Stop 停止服务。
func (srv *Service) Stop() {
	srv.locker.Lock()
	defer srv.locker.Unlock()

	if srv.cancelFunc != nil {
		srv.cancelFunc()
		srv.cancelFunc = nil
	}

	srv.state = ServiceStop
}

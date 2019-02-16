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
type ServiceFunc func(ctx context.Context) error

// ServiceState 服务的状态值
type ServiceState int8

// 几种可能的状态值
const (
	ServiceStop    ServiceState = iota + 1 // 当前处理停止状态，默认状态
	ServiceRunning                         // 正在运行
	ServiceFaild                           // 出错，不再执行后续操作
)

// ErrorHandling 出错时的处理方式
type ErrorHandling int8

// 定义几种出错时的处理方式
const (
	ContinueOnError ErrorHandling = iota + 1
	ExitOnError
)

// Service 服务模型
type Service struct {
	ID            int // 唯一标志
	Title         string
	ErrorHandling ErrorHandling

	state      ServiceState
	f          ServiceFunc
	cancelFunc context.CancelFunc
	locker     sync.Mutex

	err error // 保存上次的出错内容
}

// AddService 添加新的服务
func (m *Module) AddService(f ServiceFunc, title string, errHandling ErrorHandling) {
	srv := &Service{
		Title:         title,
		ErrorHandling: errHandling,
		state:         ServiceStop,
		f:             f,
	}

	m.Services = append(m.Services, srv)
}

// State 获取当前服务的状态
func (srv *Service) State() ServiceState {
	return srv.state
}

// Run 开始执行该服务
func (srv *Service) Run() {
	srv.locker.Lock()
	defer srv.locker.Unlock()

	if srv.state != ServiceStop {
		srv.err = fmt.Errorf("当前状态 %v 无法再次启动该服务", srv.state)
	}

	go srv.serve()
}

func (srv *Service) serve() {
	defer func() {
		if msg := recover(); msg != nil {
			srv.err = fmt.Errorf("panic:%v", msg)

			switch srv.ErrorHandling {
			case ContinueOnError:
				srv.state = ServiceStop
			case ExitOnError:
				srv.state = ServiceFaild
			}
		}
	}()

	ctx := context.Background()
	ctx, srv.cancelFunc = context.WithCancel(ctx)
	srv.state = ServiceRunning

	err := srv.f(ctx)
	if err != nil && err != context.Canceled {
		srv.err = err

		switch srv.ErrorHandling {
		case ContinueOnError:
			srv.state = ServiceStop
		case ExitOnError:
			srv.state = ServiceFaild
		}
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

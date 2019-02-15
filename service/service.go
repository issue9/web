// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package service 管理常驻后台的服务
package service

import (
	"context"
	"fmt"
	"sync"
)

// Service 服务模型
type Service struct {
	id          int // 唯一标志
	description string

	state      State
	task       TaskFunc
	cancelFunc context.CancelFunc
	locker     sync.Mutex

	err         error // 保存上次的出错内容
	errHandling ErrorHandling
}

// ID 唯一标志
func (srv *Service) ID() int {
	return srv.id
}

// Description 描述信息
func (srv *Service) Description() string {
	return srv.description
}

// State 状态值
func (srv *Service) State() State {
	return srv.state
}

// Run 开始执行该服务
func (srv *Service) Run() {
	srv.locker.Lock()
	defer srv.locker.Unlock()

	if srv.state != StateStop {
		srv.err = fmt.Errorf("当前状态 %s 无法再次启动该服务", srv.state)
	}

	go srv.serve()
}

func (srv *Service) serve() {
	defer func() {
		if msg := recover(); msg != nil {
			srv.err = fmt.Errorf("panic:%v", msg)

			switch srv.errHandling {
			case ContinueOnError:
				srv.state = StateStop
			case ExitOnError:
				srv.state = StateFaild
			}
		}
	}()

	ctx := context.Background()
	ctx, srv.cancelFunc = context.WithCancel(ctx)
	srv.state = StateRunning

	err := srv.task(ctx)
	if err != nil && err != context.Canceled {
		srv.err = err

		switch srv.errHandling {
		case ContinueOnError:
			srv.state = StateStop
		case ExitOnError:
			srv.state = StateFaild
		}
		return
	}

	srv.state = StateStop
}

// Stop 停止服务。
func (srv *Service) Stop() {
	srv.locker.Lock()
	defer srv.locker.Unlock()

	if srv.cancelFunc != nil {
		srv.cancelFunc()
		srv.cancelFunc = nil
	}

	srv.state = StateStop
}

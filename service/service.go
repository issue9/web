// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package service 管理常驻后台的服务
package service

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Service 服务模型
type Service struct {
	id          int64 // 唯一标志
	description string
	count       int // 执行次数

	state State
	next  func() chan time.Time
	task  TaskFunc

	closed     chan struct{}
	cancelFunc context.CancelFunc

	err         error // 保存上次的出错内容
	errHandling ErrorHandling
}

// ID 唯一标志
func (srv *Service) ID() int64 {
	return srv.id
}

// Description 描述信息
func (srv *Service) Description() string {
	return srv.description
}

// Count 执行次数
func (srv *Service) Count() int {
	return srv.count
}

// Run 开始执行该服务
func (srv *Service) Run() {
	if srv.state != StateWating {
		srv.err = errors.New("判断不正确，无法启动服务！")
	}

	if srv.next == nil {
		go srv.serve(time.Now())
		return
	}

	go func() {
		for now := range srv.next() {
			select {
			case <-srv.closed:
				return
			default:
				go srv.serve(now)
			}
		}
	}()
}

func (srv *Service) serve(now time.Time) {
	defer func() {
		if msg := recover(); msg != nil {
			srv.err = fmt.Errorf("panic:%v", msg)

			switch srv.errHandling {
			case ContinueOnError:
				srv.state = StateWating
			case ExitOnError:
				srv.state = StateFaild
			}
		}
	}()

	ctx := context.Background()
	ctx, srv.cancelFunc = context.WithCancel(ctx)
	srv.state = StateRunning
	if err := srv.task(ctx, now); err != nil {
		srv.err = err

		switch srv.errHandling {
		case ContinueOnError:
			srv.state = StateWating
		case ExitOnError:
			srv.state = StateFaild
		}
	}
}

// Stop 停止服务，即使定时服务，后续的也将不再执行。
func (srv *Service) Stop() {
	srv.stop()
	srv.state = StateStop
}

// Pause 停止服务，如果是定时起动的，下次依然可以执行。
func (srv *Service) Pause() {
	srv.stop()
	srv.state = StateWating
}

func (srv *Service) stop() {
	srv.closed <- struct{}{}

	if srv.cancelFunc != nil {
		srv.cancelFunc()
		srv.cancelFunc = nil
	}
}

// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package service

import (
	"time"

	"github.com/issue9/autoinc"
)

// Services 服务管理
type Services struct {
	services []*Service
	ai       *autoinc.AutoInc
}

// NewServices 声明新的 Services 变量
func NewServices() *Services {
	return &Services{
		services: make([]*Service, 10),
		ai:       autoinc.New(1, 1, 100),
	}
}

// New 添加新的服务
//
// next 表示下次执行此服务的时间，如果是一个一次性的常驻服务，请使用 nil 代替。
//
// NOTE: 如果为服务生成唯一 ID 失败，则会 panic。
func (s *Services) New(task TaskFunc, description string, errHandling ErrorHandling, next func() chan time.Time) {
	srv := &Service{
		id:          s.ai.MustID(),
		description: description,
		state:       StateWating,
		task:        task,
		next:        next,
		closed:      make(chan struct{}, 1),
		errHandling: errHandling,
	}

	srv.Run()
	s.services = append(s.services, srv)
}

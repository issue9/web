// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package service

import "time"

// Services 服务管理
type Services struct {
	services []*Service
}

// NewServices 声明新的 Services 变量
func NewServices() *Services {
	return &Services{
		services: make([]*Service, 10),
	}
}

// New 添加新的服务
func (s *Services) New(task TaskFunc, description string, errHandling ErrorHandling) {
	srv := &Service{
		id:          "",
		description: description,
		state:       StateWating,
		task:        task,
		closed:      make(chan struct{}, 1),
		errHandling: errHandling,
	}

	srv.Run()
	s.services = append(s.services, srv)
}

// NewTicker 添加定时启动的服务
func (s *Services) NewTicker(task TaskFunc, next func() chan time.Time, description string, errHandling ErrorHandling) {
	srv := &Service{
		id:          "",
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

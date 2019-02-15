// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package service

import "sync"

// Services 服务管理
// NOTE: 最多只能管理 math.MaxInt64 个服务。
type Services struct {
	services []*Service
	locker   sync.Mutex
}

// NewServices 声明新的 Services 变量
func NewServices() *Services {
	return &Services{
		services: make([]*Service, 10),
	}
}

// New 添加新的服务
func (s *Services) New(task TaskFunc, description string, errHandling ErrorHandling) {
	s.locker.Lock()
	defer s.locker.Unlock()

	srv := &Service{
		id:          len(s.services) + 1, // 从 1 开始计数
		description: description,
		state:       StateStop,
		task:        task,
		errHandling: errHandling,
	}

	s.services = append(s.services, srv)
}

// Serve 依次启动服务
func (s *Services) Serve() {
	for _, srv := range s.services {
		srv.Run()
	}
}

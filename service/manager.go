// SPDX-License-Identifier: MIT

package service

import (
	"context"
	"time"

	"github.com/issue9/logs/v3"
	"github.com/issue9/scheduled"
)

// Manager 服务管理
type Manager struct {
	services  []*Service
	scheduled *scheduled.Server
	logs      *logs.Logs
	running   bool
}

// NewManager 返回 Manager
func NewManager(loc *time.Location, logs *logs.Logs) *Manager {
	mgr := &Manager{
		services:  make([]*Service, 0, 100),
		scheduled: scheduled.NewServer(loc),
		logs:      logs,
	}

	mgr.AddService("计划任务", func(ctx context.Context) error {
		go func() {
			if err := mgr.Scheduled().Serve(logs.ERROR(), logs.DEBUG()); err != nil {
				logs.Error(err)
			}
		}()

		<-ctx.Done()
		mgr.scheduled.Stop()
		return context.Canceled
	})

	return mgr
}

func (mgr *Manager) Scheduled() *scheduled.Server { return mgr.scheduled }

// Run 运行所有服务
func (mgr *Manager) Run() {
	if mgr.running {
		panic("服务已经在运行")
	}

	for _, s := range mgr.services {
		s.Run()
	}

	mgr.running = true
}

// Stop 停止所有服务
func (mgr *Manager) Stop() {
	for _, s := range mgr.services {
		s.Stop()
	}
	mgr.running = false
}

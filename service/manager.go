// SPDX-License-Identifier: MIT

package service

import (
	"time"

	"github.com/issue9/logs/v3"
	"github.com/issue9/scheduled"
)

// Manager 服务管理
type Manager struct {
	services  []*Service
	scheduled *scheduled.Server
	logs      *logs.Logs

	running bool
}

// NewManager 返回 Manager
func NewManager(logs *logs.Logs, loc *time.Location) *Manager {
	mgr := &Manager{
		services:  make([]*Service, 0, 100),
		scheduled: scheduled.NewServer(loc, logs.ERROR(), logs.DEBUG()),
		logs:      logs,
	}

	mgr.AddService("计划任务", mgr.scheduledService)

	return mgr
}

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

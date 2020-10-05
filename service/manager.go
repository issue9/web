// SPDX-License-Identifier: MIT

package service

// Manager 服务管理
type Manager struct {
	services []*Service
}

// NewManager 返回 Manager
func NewManager() *Manager {
	return &Manager{
		services: make([]*Service, 0, 100),
	}
}

// Run 运行所有服务
func (mgr *Manager) Run() {
	for _, s := range mgr.services {
		s.Run()
	}
}

// Stop 停止所有服务
func (mgr *Manager) Stop() {
	for _, s := range mgr.services {
		s.Stop()
	}
}

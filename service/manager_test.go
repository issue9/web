// SPDX-License-Identifier: MIT

package service

import (
	"testing"
	"time"

	"github.com/issue9/assert"
	"github.com/issue9/logs/v3"
)

func newManager(a *assert.Assertion, t *time.Location) *Manager {
	l, err := logs.New(nil)
	a.NotError(err).NotNil(l)

	mgr := NewManager(l, t)
	a.NotNil(mgr)
	return mgr
}

func TestManager(t *testing.T) {
	a := assert.New(t)
	mgr := newManager(a, time.Local)
	a.NotNil(mgr)
	a.Equal(1, len(mgr.Services())) // 默认的计划任务

	// 未运行

	a.False(mgr.running)
	srv0 := mgr.services[0]
	a.Equal(srv0.State(), Stopped)

	s1, start1, exit1 := buildSrv1()
	mgr.AddService("srv1", s1)
	a.Equal(2, len(mgr.Services()))
	srv1 := mgr.services[1]
	time.Sleep(500 * time.Microsecond) // 等待主服务设置状态值
	a.Equal(srv1.f, s1)                // 并不会改变状态

	// 运行中

	mgr.Run()
	<-start1
	time.Sleep(500 * time.Microsecond) // 等待主服务设置状态值
	a.Equal(Running, srv0.State()).
		Equal(Running, srv1.State())

	a.Panic(func() { mgr.Run() })

	// 运行中添加
	s2, start2, exit2 := buildSrv1()
	mgr.AddService("srv2", s2)
	a.Equal(3, len(mgr.Services()))
	srv2 := mgr.services[2]
	<-start2
	time.Sleep(500 * time.Microsecond) // 等待主服务设置状态值
	a.Equal(Running, srv2.State())     // 运行中添加自动运行服务

	mgr.Stop()
	<-exit1
	<-exit2
	time.Sleep(500 * time.Microsecond) // 等待主服务设置状态值
	a.False(mgr.running)
	a.Equal(srv1.State(), Stopped)
	a.Equal(srv0.State(), Stopped)
	a.Equal(srv2.State(), Stopped)
}

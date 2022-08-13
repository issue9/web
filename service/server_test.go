// SPDX-License-Identifier: MIT

package service

import (
	"testing"
	"time"

	"github.com/issue9/assert/v3"
	"github.com/issue9/logs/v4"
	"golang.org/x/text/language"

	"github.com/issue9/web/internal/locale"
)

func newServer(a *assert.Assertion) *Server {
	a.TB().Helper()
	srv := InternalNewServer(logs.New(nil), locale.New(time.Local, language.SimplifiedChinese))
	a.NotNil(srv)
	return srv
}

func TestServer_service(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a)

	// 未运行

	a.False(srv.Running())
	a.Equal(0, len(srv.Services()))

	s1, start1, exit1 := buildService()
	srv.Add("srv1", s1)
	a.Equal(1, len(srv.Services()))
	srv1 := srv.services[0]
	a.Equal(srv1.f, s1) // 并不会改变状态
	a.Equal(srv1.State(), Stopped)

	// 运行中

	srv.Run()
	a.Equal(2, len(srv.Services())) // 在 runServices 中添加了 Scheduled 服务
	srv.running = true
	<-start1
	time.Sleep(500 * time.Microsecond) // 等待主服务设置状态值
	sched := srv.services[1]
	a.Equal(Running, sched.State()).
		Equal(Running, srv1.State())

	// 运行中添加
	s2, start2, exit2 := buildService()
	srv.Add("srv2", s2)
	a.Equal(3, len(srv.Services()))
	srv2 := srv.services[2]
	<-start2
	time.Sleep(500 * time.Microsecond) // 等待主服务设置状态值
	a.Equal(Running, srv2.State())     // 运行中添加自动运行服务

	srv.Stop()
	<-exit1
	<-exit2
	time.Sleep(500 * time.Microsecond) // 等待主服务设置状态值
	a.Equal(srv1.State(), Stopped)
	a.Equal(sched.State(), Stopped)
	a.Equal(srv2.State(), Stopped)
}

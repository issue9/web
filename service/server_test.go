// SPDX-License-Identifier: MIT

package service

import (
	"testing"
	"time"

	"github.com/issue9/assert/v3"
	"github.com/issue9/localeutil"
	"github.com/issue9/scheduled"

	"github.com/issue9/web/logs"
)

var _ Services = &Server{}

func newServer(a *assert.Assertion) *Server {
	a.TB().Helper()
	srv := NewServer(time.Local, logs.New(nil, nil))
	a.NotNil(srv)
	return srv
}

func TestServer_service(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a)

	// 未运行

	a.False(srv.running)
	a.Equal(1, len(srv.Services()))

	s1, start1, exit1 := buildService()
	srv.Add(localeutil.Phrase("srv1"), s1)
	a.Equal(2, len(srv.Services()))
	sched := srv.services[0]
	srv1 := srv.services[1]
	a.Equal(srv1.service, s1) // 并不会改变状态
	a.Equal(srv1.State(), scheduled.Stopped).
		Equal(sched.State(), scheduled.Stopped)

	// 运行中

	srv.Run()
	a.Equal(2, len(srv.Services()))
	srv.running = true
	<-start1
	time.Sleep(500 * time.Microsecond) // 等待主服务设置状态值
	a.Equal(scheduled.Running, sched.State()).
		Equal(scheduled.Running, srv1.State())

	// 运行中添加
	s2, start2, exit2 := buildService()
	srv.Add(localeutil.Phrase("srv2"), s2)
	a.Equal(3, len(srv.Services()))
	srv2 := srv.services[2]
	<-start2
	time.Sleep(500 * time.Microsecond)       // 等待主服务设置状态值
	a.Equal(scheduled.Running, srv2.State()) // 运行中添加自动运行服务

	srv.Stop()
	<-exit1
	<-exit2
	time.Sleep(500 * time.Microsecond) // 等待主服务设置状态值
	a.Equal(srv1.State(), scheduled.Stopped)
	a.Equal(sched.State(), scheduled.Stopped)
	a.Equal(srv2.State(), scheduled.Stopped)
}

func TestServer_scheduled(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a)
	a.Equal(0, len(srv.Jobs()))

	srv.scheduled.At("at", func(t time.Time) error {
		println("at:", t.Format(time.RFC3339))
		return nil
	}, time.Now(), false)
	a.Equal(1, len(srv.scheduled.Jobs()))
}

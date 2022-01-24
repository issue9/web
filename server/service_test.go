// SPDX-License-Identifier: MIT

package server

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/issue9/assert/v2"
)

const (
	tickTimer  = 500 * time.Microsecond
	panicTimer = 50 * tickTimer // windows 下此值不能过小，否则测试容易出错
)

func buildSrv1() (f ServiceFunc, start, exit chan struct{}) {
	exit = make(chan struct{}, 1)
	start = make(chan struct{}, 1)

	return func(ctx context.Context) error {
		defer func() {
			exit <- struct{}{}
		}()

		inited := false
		for now := range time.Tick(tickTimer) {
			select {
			case <-ctx.Done():
				fmt.Println("cancel srv1")
				return ctx.Err()
			default:
				fmt.Println("srv1:", now)
				if !inited {
					inited = true
					start <- struct{}{}
				}
			}
		}
		return nil
	}, start, exit
}

// panic
func buildSrv2() (f ServiceFunc, start, exit chan struct{}) {
	exit = make(chan struct{}, 1)
	start = make(chan struct{}, 1)

	return func(ctx context.Context) error {
		defer func() {
			exit <- struct{}{}
		}()

		inited := false
		timer := time.NewTimer(panicTimer)
		for now := range time.Tick(tickTimer) {
			select {
			case <-ctx.Done():
				fmt.Println("cancel srv2")
				return ctx.Err()
			case <-timer.C:
				fmt.Println("panic srv2")
				panic("panic srv2")
			default:
				if !inited {
					inited = true
					start <- struct{}{}
				}
				fmt.Println("srv2:", now)
			}
		}
		return nil
	}, start, exit
}

// error
func buildSrv3() (f ServiceFunc, start, exit chan struct{}) {
	exit = make(chan struct{}, 1)
	start = make(chan struct{}, 1)

	return func(ctx context.Context) error {
		defer func() {
			exit <- struct{}{}
		}()

		inited := false
		timer := time.NewTimer(panicTimer)
		for now := range time.Tick(tickTimer) {
			select {
			case <-ctx.Done():
				fmt.Println("cancel srv3")
				return ctx.Err()
			case <-timer.C:
				fmt.Println("panic srv2")
				return errors.New("error")
			default:
				fmt.Println("srv3:", now)
				if !inited {
					inited = true
					start <- struct{}{}
				}
			}
		}
		return nil
	}, start, exit
}

func TestService_srv1(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, &Options{Location: time.Local})
	defer srv.stopServices()

	srv1, start, exit := buildSrv1()
	srv.AddService("srv1", srv1)
	srv.runServices()
	srv.serving = true
	<-start
	time.Sleep(500 * time.Microsecond) // 等待主服务设置状态值
	s1 := srv.services[0]
	time.Sleep(500 * time.Microsecond) // 等待主服务设置状态值
	a.Equal(s1.State(), ServiceRunning)
	s1.Stop()
	<-exit
	time.Sleep(500 * time.Microsecond) // 等待主服务设置状态值
	a.Equal(s1.State(), ServiceStopped)

	s1.Run()
	s1.Run() // 在运行状态再次运行，不启作用
	<-start
	time.Sleep(500 * time.Microsecond) // 等待主服务设置状态值
	a.Equal(s1.State(), ServiceRunning)
	s1.Stop()
	<-exit
	time.Sleep(500 * time.Microsecond) // 等待主服务设置状态值
	a.Equal(s1.State(), ServiceStopped)
}

func TestService_srv2(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, &Options{Location: time.Local})
	defer srv.stopServices()

	srv2, start, exit := buildSrv2()
	srv.AddService("srv2", srv2)
	srv.runServices() // 注册并运行服务
	srv.serving = true
	s2 := srv.services[0]
	<-start
	time.Sleep(500 * time.Microsecond) // 等待主服务设置状态值
	a.Equal(s2.State(), ServiceRunning)
	s2.Stop()
	<-exit
	time.Sleep(500 * time.Microsecond) // 等待主服务设置状态值
	a.Equal(s2.State(), ServiceStopped)

	// 再次运行，等待 panic
	s2.Run()
	<-start
	<-exit
	time.Sleep(500 * time.Microsecond) // 等待主服务设置状态值
	a.Equal(s2.State(), ServiceFailed)
	a.NotEmpty(s2.Err())

	// 出错后，还能正确运行和结束
	s2.Run()
	<-start
	time.Sleep(500 * time.Microsecond) // 等待主服务设置状态值
	a.Equal(s2.State(), ServiceRunning)
	s2.Stop()
	<-exit
	time.Sleep(500 * time.Microsecond) // 等待主服务设置状态值
	a.Equal(s2.State(), ServiceStopped)
}

func TestService_srv3(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, &Options{Location: time.Local})
	defer srv.stopServices()

	srv3, start, exit := buildSrv3()
	srv.AddService("srv3", srv3)
	srv.runServices()
	srv.serving = true
	s3 := srv.services[0]
	<-start
	time.Sleep(500 * time.Microsecond) // 等待主服务设置状态值
	a.Equal(s3.State(), ServiceRunning)

	<-exit                             // 等待超过返回错误
	time.Sleep(500 * time.Microsecond) // 等待主服务设置状态值
	a.Equal(s3.State(), ServiceFailed)
	a.NotNil(s3.Err())

	// 再次运行
	s3.Run()
	<-start
	time.Sleep(500 * time.Microsecond) // 等待主服务设置状态值
	a.Equal(s3.State(), ServiceRunning)
	s3.Stop()
	<-exit
	time.Sleep(500 * time.Microsecond) // 等待主服务设置状态值
	a.Equal(s3.State(), ServiceStopped)
}

func TestServer_service(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, nil)

	// 未运行

	a.False(srv.Serving())
	a.Equal(0, len(srv.Services()))

	s1, start1, exit1 := buildSrv1()
	srv.AddService("srv1", s1)
	a.Equal(1, len(srv.Services()))
	srv1 := srv.services[0]
	a.Equal(srv1.f, s1) // 并不会改变状态
	a.Equal(srv1.State(), ServiceStopped)

	// 运行中

	srv.runServices()
	a.Equal(2, len(srv.Services())) // 在 runServices 中添加了 Scheduled 服务
	srv.serving = true
	<-start1
	time.Sleep(500 * time.Microsecond) // 等待主服务设置状态值
	sched := srv.services[1]
	a.Equal(ServiceRunning, sched.State()).
		Equal(ServiceRunning, srv1.State())

	// 运行中添加
	s2, start2, exit2 := buildSrv1()
	srv.AddService("srv2", s2)
	a.Equal(3, len(srv.Services()))
	srv2 := srv.services[2]
	<-start2
	time.Sleep(500 * time.Microsecond)    // 等待主服务设置状态值
	a.Equal(ServiceRunning, srv2.State()) // 运行中添加自动运行服务

	srv.stopServices()
	<-exit1
	<-exit2
	time.Sleep(500 * time.Microsecond) // 等待主服务设置状态值
	a.Equal(srv1.State(), ServiceStopped)
	a.Equal(sched.State(), ServiceStopped)
	a.Equal(srv2.State(), ServiceStopped)
}

func TestServer_scheduled(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, nil)

	a.Equal(0, len(srv.Jobs()))

	srv.scheduled.At("at", func(t time.Time) error {
		println("at:", t.Format(time.RFC3339))
		return nil
	}, time.Now(), false)
	a.Equal(1, len(srv.scheduled.Jobs()))
}

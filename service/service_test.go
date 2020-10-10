// SPDX-License-Identifier: MIT

package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/issue9/assert"
)

const (
	tickTimer  = 500 * time.Microsecond
	panicTimer = 50 * tickTimer // windows 下此值不能过小，否则测试容易出错
)

func buildSrv1() (f Func, start, exit chan struct{}) {
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
func buildSrv2() (f Func, start, exit chan struct{}) {
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
func buildSrv3() (f Func, start, exit chan struct{}) {
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
	a := assert.New(t)
	srv := NewManager()

	srv1, start, exit := buildSrv1()
	srv.AddService(srv1, "srv1")
	srv.Run()
	<-start
	time.Sleep(500 * time.Microsecond) // 等待主服务设置状态值
	a.Equal(1, len(srv.services))
	s1 := srv.services[0]
	time.Sleep(500 * time.Microsecond) // 等待主服务设置状态值
	a.Equal(s1.State(), Running)
	s1.Stop()
	<-exit
	time.Sleep(500 * time.Microsecond) // 等待主服务设置状态值
	a.Equal(s1.State(), Stopped)

	s1.Run()
	s1.Run() // 在运行状态再次运行，不启作用
	<-start
	time.Sleep(500 * time.Microsecond) // 等待主服务设置状态值
	a.Equal(s1.State(), Running)
	s1.Stop()
	<-exit
	time.Sleep(500 * time.Microsecond) // 等待主服务设置状态值
	a.Equal(s1.State(), Stopped)
}

func TestService_srv2(t *testing.T) {
	a := assert.New(t)
	srv := NewManager()

	srv2, start, exit := buildSrv2()
	srv.AddService(srv2, "srv2")
	srv.Run() // 注册并运行服务
	s2 := srv.services[0]
	<-start
	time.Sleep(500 * time.Microsecond) // 等待主服务设置状态值
	a.Equal(s2.State(), Running)
	s2.Stop()
	<-exit
	time.Sleep(500 * time.Microsecond) // 等待主服务设置状态值
	a.Equal(s2.State(), Stopped)

	// 再次运行，等待 panic
	s2.Run()
	<-start
	<-exit
	time.Sleep(500 * time.Microsecond) // 等待主服务设置状态值
	a.Equal(s2.State(), Failed)
	a.NotEmpty(s2.Err())

	// 出错后，还能正确运行和结束
	s2.Run()
	<-start
	time.Sleep(500 * time.Microsecond) // 等待主服务设置状态值
	a.Equal(s2.State(), Running)
	s2.Stop()
	<-exit
	time.Sleep(500 * time.Microsecond) // 等待主服务设置状态值
	a.Equal(s2.State(), Stopped)
}

func TestService_srv3(t *testing.T) {
	a := assert.New(t)
	srv := NewManager()

	srv3, start, exit := buildSrv3()
	srv.AddService(srv3, "srv3")
	srv.Run()
	s3 := srv.services[0]
	<-start
	time.Sleep(500 * time.Microsecond) // 等待主服务设置状态值
	a.Equal(s3.State(), Running)

	<-exit                             // 等待超过返回错误
	time.Sleep(500 * time.Microsecond) // 等待主服务设置状态值
	a.Equal(s3.State(), Failed)
	a.NotNil(s3.Err())

	// 再次运行
	s3.Run()
	<-start
	time.Sleep(500 * time.Microsecond) // 等待主服务设置状态值
	a.Equal(s3.State(), Running)
	s3.Stop()
	<-exit
	time.Sleep(500 * time.Microsecond) // 等待主服务设置状态值
	a.Equal(s3.State(), Stopped)
}

func TestService_String(t *testing.T) {
	a := assert.New(t)

	var state State
	a.Equal(state.String(), "stopped")

	a.Equal(Failed.String(), "failed")
	a.Equal(Running.String(), "running")
	a.Equal(Stopped.String(), "stopped")

	state = -1
	a.Equal(state.String(), "<unknown>")
}

// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package module

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/issue9/assert"

	"github.com/issue9/web/internal/webconfig"
)

const (
	tickTimer  = 500 * time.Microsecond
	panicTimer = 5 * tickTimer
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
		for now := range time.Tick(tickTimer) {
			select {
			case <-ctx.Done():
				fmt.Println("cancel srv3")
				return ctx.Err()
			default:
				fmt.Println("srv3:", now)
				if !inited {
					inited = true
					start <- struct{}{}
				} else {
					return errors.New("Error")
				}
			}
		}
		return nil
	}, start, exit
}

func TestModule_AddService(t *testing.T) {
	a := assert.New(t)
	ms, err := NewModules(&webconfig.WebConfig{})
	a.NotError(err).NotNil(ms)
	m := newModule(ms, "m1", "m1 desc")
	a.NotNil(m)

	srv1 := func(ctx context.Context) error { return nil }

	ml := len(m.inits)
	m.AddService(srv1, "srv1")
	a.Equal(ml+1, len(m.inits))
}

func TestService_srv1(t *testing.T) {
	a := assert.New(t)
	ms, err := NewModules(&webconfig.WebConfig{})
	a.NotError(err).NotNil(ms)

	m := ms.NewModule("m1", "m1 desc")
	a.NotNil(m)
	a.Empty(ms.services)

	srv1, start, exit := buildSrv1()
	m.AddService(srv1, "srv1")
	a.NotError(ms.Init("", log.New(os.Stdout, "", 0))) // 注册并运行服务
	<-start
	time.Sleep(20 * time.Microsecond) // 等待其它内容初始化完成
	a.Equal(1, len(ms.services))
	s1 := ms.services[0]
	a.Equal(s1.Module, m)
	a.Equal(s1.State(), ServiceRunning)
	s1.Stop()
	<-exit
	a.Equal(s1.State(), ServiceStop)

	s1.Run()
	s1.Run() // 在运行状态再次运行，不启作用
	<-start
	a.Equal(s1.State(), ServiceRunning)
	s1.Stop()
	<-exit
	a.Equal(s1.State(), ServiceStop)
}

func TestService_srv2(t *testing.T) {
	a := assert.New(t)
	ms, err := NewModules(&webconfig.WebConfig{})
	a.NotError(err).NotNil(ms)

	m := ms.NewModule("m1", "m1 desc")
	a.NotNil(m)
	a.Empty(ms.services)

	srv2, start, exit := buildSrv2()
	m.AddService(srv2, "srv2")
	a.NotError(ms.Init("", nil)) // 注册并运行服务
	s2 := ms.services[0]
	<-start
	time.Sleep(20 * time.Microsecond) // 等待服务启动完成
	a.Equal(s2.State(), ServiceRunning)
	s2.Stop()
	<-exit
	a.Equal(s2.State(), ServiceStop)

	// 再次运行，等待 panic
	s2.Run()
	<-start
	<-exit
	a.Equal(s2.State(), ServiceFailed)
	a.NotEmpty(s2.Err())

	// 出错后，还能正确运行和结束
	s2.Run()
	<-start
	a.Equal(s2.State(), ServiceRunning)
	s2.Stop()
	<-exit
	a.Equal(s2.State(), ServiceStop)
}

func TestService_srv3(t *testing.T) {
	a := assert.New(t)
	ms, err := NewModules(&webconfig.WebConfig{})
	a.NotError(err).NotNil(ms)

	m := ms.NewModule("m1", "m1 desc")
	a.NotNil(m)
	a.Empty(ms.services)

	srv3, start, exit := buildSrv3()
	m.AddService(srv3, "srv3")
	a.NotError(ms.Init("", nil)) // 注册并运行服务
	s3 := ms.services[0]
	<-start
	time.Sleep(20 * time.Microsecond) // 等待服务启动完成
	a.Equal(s3.State(), ServiceRunning)

	<-exit // 等待超过返回错误
	a.Equal(s3.State(), ServiceFailed)
	a.NotNil(s3.Err())

	// 再次运行
	s3.Run()
	<-start
	a.Equal(s3.State(), ServiceRunning)
	s3.Stop()
	<-exit
	a.Equal(s3.State(), ServiceStop)
}

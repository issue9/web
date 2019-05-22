// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

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
	a := assert.New(t)
	app := newApp(a)

	srv1, start, exit := buildSrv1()
	app.AddService(srv1, "srv1")
	app.runServices()
	<-start
	a.Equal(2, len(app.services)) // 自带一个 scheduled
	s1 := app.services[1]         // 0 为 scheduled
	a.Equal(s1.State(), ServiceRunning)
	s1.Stop()
	<-exit
	a.Equal(s1.State(), ServiceStoped)

	s1.Run()
	s1.Run() // 在运行状态再次运行，不启作用
	<-start
	a.Equal(s1.State(), ServiceRunning)
	s1.Stop()
	<-exit
	a.Equal(s1.State(), ServiceStoped)
}

func TestService_srv2(t *testing.T) {
	a := assert.New(t)
	app := newApp(a)

	srv2, start, exit := buildSrv2()
	app.AddService(srv2, "srv2")
	app.runServices()     // 注册并运行服务
	s2 := app.services[1] // 0 为 scheduled
	<-start
	a.Equal(s2.State(), ServiceRunning)
	s2.Stop()
	<-exit
	a.Equal(s2.State(), ServiceStoped)

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
	a.Equal(s2.State(), ServiceStoped)
}

func TestService_srv3(t *testing.T) {
	a := assert.New(t)
	app := newApp(a)

	srv3, start, exit := buildSrv3()
	app.AddService(srv3, "srv3")
	app.runServices()
	s3 := app.services[1] // 0 为 scheduled
	<-start
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
	a.Equal(s3.State(), ServiceStoped)
}

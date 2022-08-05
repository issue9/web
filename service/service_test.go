// SPDX-License-Identifier: MIT

package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/issue9/assert/v2"
)

const tickTimer = 500 * time.Microsecond

func buildService() (f ServiceFunc, start, exit chan struct{}) {
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
				fmt.Println("cancel service")
				return ctx.Err()
			default:
				fmt.Println("service:", now)
				if !inited {
					inited = true
					start <- struct{}{}
				}
			}
		}
		return nil
	}, start, exit
}

func buildPanicService() (f ServiceFunc, start, exit chan struct{}) {
	exit = make(chan struct{}, 1)
	start = make(chan struct{}, 1)

	return func(ctx context.Context) error {
		defer func() {
			exit <- struct{}{}
		}()

		count := 0
		p := make(chan struct{}, 1)
		for now := range time.Tick(tickTimer) {
			select {
			case <-ctx.Done():
				fmt.Println("cancel panic service")
				return ctx.Err()
			case <-p:
				fmt.Println("enter panic service")
				panic("service panic")
			default:
				if count == 0 {
					start <- struct{}{}
				}
				count++
				fmt.Println("panic service:", now)
				if count >= 4 {
					p <- struct{}{}
				}
			}
		}
		return nil
	}, start, exit
}

func buildErrorService() (f ServiceFunc, start, exit chan struct{}) {
	exit = make(chan struct{}, 1)
	start = make(chan struct{}, 1)

	return func(ctx context.Context) error {
		defer func() {
			exit <- struct{}{}
		}()

		count := 0
		p := make(chan struct{}, 1)
		for now := range time.Tick(tickTimer) {
			select {
			case <-ctx.Done():
				fmt.Println("cancel error service")
				return ctx.Err()
			case <-p:
				fmt.Println("enter error service")
				return errors.New("service error")
			default:
				if count == 0 {
					start <- struct{}{}
				}
				count++
				fmt.Println("error service:", now)
				if count >= 4 {
					p <- struct{}{}
				}
			}
		}
		return nil
	}, start, exit
}

func TestService_service(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a)
	defer s.Stop()

	srv1, start, exit := buildService()
	s.Add("srv1", srv1)
	s.Run()
	s.running = true
	<-start
	time.Sleep(500 * time.Microsecond) // 等待主服务设置状态值
	s1 := s.services[0]
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

func TestService_panic(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a)
	defer s.Stop()

	srv2, start, exit := buildPanicService()
	s.Add("srv2", srv2)
	s.Run() // 注册并运行服务
	s.running = true
	s2 := s.services[0]
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

func TestService_error(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a)
	defer s.Stop()

	srv3, start, exit := buildErrorService()
	s.Add("srv3", srv3)
	s.Run()
	s.running = true
	s3 := s.services[0]
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

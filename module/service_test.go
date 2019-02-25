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

var (
	// 正常服务
	srv1 = func(ctx context.Context) error {
		for now := range time.Tick(tickTimer) {
			select {
			case <-ctx.Done():
				fmt.Println("cancel srv1")
				return ctx.Err()
			default:
				fmt.Println("srv1:", now)
			}
		}
		return nil
	}

	// panic
	srv2 = func(ctx context.Context) error {
		timer := time.NewTimer(panicTimer)

		for now := range time.Tick(tickTimer) {
			select {
			case <-ctx.Done():
				fmt.Println("cancel srv2")
				return ctx.Err()
			case <-timer.C:
				panic("panic srv2")
			default:
				fmt.Println("srv2:", now)
			}
		}
		return nil
	}

	// error
	srv3 = func(ctx context.Context) error {
		for now := range time.Tick(tickTimer) {
			select {
			case <-ctx.Done():
				fmt.Println("cancel srv3")
				return ctx.Err()
			default:
				fmt.Println("srv3:", now)
				return errors.New("Error")
			}
		}
		return nil
	}
)

func stopService(a *assert.Assertion, srv *Service) {
	srv.Stop()
	time.Sleep(3 * tickTimer) // 等待停止
	a.Equal(srv.State(), ServiceStop)
}

func TestModule_AddService(t *testing.T) {
	a := assert.New(t)
	ms, err := NewModules(&webconfig.WebConfig{})
	a.NotError(err).NotNil(ms)
	m := newModule(ms, "m1", "m1 desc")
	a.NotNil(m)

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

	m.AddService(srv1, "srv1")
	a.NotError(ms.Init("", log.New(os.Stdout, "", 0))) // 注册并运行服务
	time.Sleep(20 * time.Microsecond)                  // 等待服务启动完成
	a.Equal(1, len(ms.services))
	srv1 := ms.services[0]
	a.Equal(srv1.Module, m)
	a.Equal(srv1.State(), ServiceRunning)
	stopService(a, srv1)

	// 再次运行
	srv1.Run()
	srv1.Run()                         // 在运行状态再次运行，不启作用
	time.Sleep(500 * time.Microsecond) // 等待服务启动完成
	a.Equal(srv1.State(), ServiceRunning)
	stopService(a, srv1)
}

func TestService_srv2(t *testing.T) {
	a := assert.New(t)
	ms, err := NewModules(&webconfig.WebConfig{})
	a.NotError(err).NotNil(ms)

	m := ms.NewModule("m1", "m1 desc")
	a.NotNil(m)
	a.Empty(ms.services)

	m.AddService(srv2, "srv2")
	a.NotError(ms.Init("", nil)) // 注册并运行服务
	srv2 := ms.services[0]
	time.Sleep(20 * time.Microsecond) // 等待服务启动完成
	a.Equal(srv2.State(), ServiceRunning)
	stopService(a, srv2)

	// 再次运行，等待 panic
	srv2.Run()
	time.Sleep(panicTimer * 2) // 等待 panic 触发
	a.Equal(srv2.State(), ServiceFailed)
	a.NotEmpty(srv2.Err())

	// 出错后，还能正确运行和结束
	srv2.Run()
	time.Sleep(20 * time.Microsecond) // 等待服务启动完成
	stopService(a, srv2)
}

func TestService_srv3(t *testing.T) {
	a := assert.New(t)
	ms, err := NewModules(&webconfig.WebConfig{})
	a.NotError(err).NotNil(ms)

	m := ms.NewModule("m1", "m1 desc")
	a.NotNil(m)
	a.Empty(ms.services)

	m.AddService(srv3, "srv3")
	a.NotError(ms.Init("", nil)) // 注册并运行服务
	srv3 := ms.services[0]
	time.Sleep(20 * time.Microsecond) // 等待服务启动完成
	a.Equal(srv3.State(), ServiceRunning)
	time.Sleep(2 * tickTimer) // 等待超过返回错误
	a.Equal(srv3.State(), ServiceFailed)
	a.NotNil(srv3.Err())

	// 再次运行
	srv3.Run()
	time.Sleep(20 * time.Microsecond) // 等待服务启动完成
	a.Equal(srv3.State(), ServiceRunning)
	stopService(a, srv3)
}

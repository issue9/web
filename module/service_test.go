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

const panicTimer = 500 * time.Millisecond

var (
	// 正常服务
	srv1 = func(ctx context.Context) error {
		for now := range time.Tick(500 * time.Microsecond) {
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

		for now := range time.Tick(500 * time.Microsecond) {
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
		for now := range time.Tick(500 * time.Microsecond) {
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

func TestModule_AddService(t *testing.T) {
	a := assert.New(t)
	ms, err := NewModules(&webconfig.WebConfig{})
	a.NotError(err).NotNil(ms)
	m := newModule(ms, TypeModule, "m1", "m1 desc")
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
	time.Sleep(400 * time.Microsecond)                 // 等待服务启动完成
	srv1 := ms.services[0]
	a.Equal(srv1.State(), ServiceRunning)
	srv1.Stop()
	a.Equal(srv1.State(), ServiceStop)
	time.Sleep(400 * time.Microsecond) // 等待停止

	// 再次运行
	srv1.Run()
	time.Sleep(400 * time.Microsecond) // 等待服务启动完成
	a.Equal(srv1.State(), ServiceRunning)
	srv1.Stop()
	a.Equal(srv1.State(), ServiceStop)
	time.Sleep(400 * time.Microsecond) // 等待停止
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
	time.Sleep(400 * time.Microsecond) // 等待服务启动完成
	a.Equal(srv2.State(), ServiceRunning)
	srv2.Stop()
	a.Equal(srv2.State(), ServiceStop)
	time.Sleep(400 * time.Microsecond) // 等待停止

	// 再次运行，等待 panic
	srv2.Run()
	time.Sleep(panicTimer * 2) // 等待 panic 触发
	a.Equal(srv2.State(), ServiceFailed)

	// 出错后，还能正确运行和结束
	srv2.Run()
	time.Sleep(400 * time.Microsecond) // 等待服务启动完成
	srv2.Stop()
	time.Sleep(400 * time.Microsecond) // 等待停止
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
	time.Sleep(400 * time.Microsecond) // 等待服务启动完成
	a.Equal(srv3.State(), ServiceRunning)
	time.Sleep(300 * time.Microsecond) // 等待超过返回错误
	a.Equal(srv3.State(), ServiceFailed)
	a.NotNil(srv3.Err())

	// 再次运行
	srv3.Run()
	time.Sleep(400 * time.Microsecond) // 等待服务启动完成
	a.Equal(srv3.State(), ServiceRunning)
	srv3.Stop()
	a.Equal(srv3.State(), ServiceStop)
	time.Sleep(400 * time.Microsecond) // 等待停止
}

// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package module

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/issue9/assert"
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
	m := New(TypeModule, "m1", "m1 desc")
	a.NotNil(m)

	a.Nil(m.services)
	m.AddService(srv1, "srv1")
	a.Equal(len(m.services), 1)
	a.Equal(m.services[0].State(), ServiceStop)
}

func TestService(t *testing.T) {
	a := assert.New(t)

	m := New(TypeModule, "m1", "m1 desc")
	a.NotNil(m)
	a.Nil(m.services)

	// srv1

	m.AddService(srv1, "srv1")
	srv1 := m.services[0]
	srv1.Run()
	time.Sleep(200 * time.Microsecond) // 等待服务启动完成
	a.Equal(srv1.State(), ServiceRunning)
	srv1.Stop()
	a.Equal(srv1.State(), ServiceStop)

	// 再次运行
	srv1.Run()
	time.Sleep(200 * time.Microsecond) // 等待服务启动完成
	a.Equal(srv1.State(), ServiceRunning)
	srv1.Stop()
	a.Equal(srv1.State(), ServiceStop)

	// srv2

	m.AddService(srv2, "srv2")
	srv2 := m.services[1]
	srv2.Run()
	time.Sleep(200 * time.Microsecond) // 等待服务启动完成
	a.Equal(srv2.State(), ServiceRunning)
	srv2.Stop()
	a.Equal(srv2.State(), ServiceStop)

	// 再次运行，等待 panic
	srv2.Run()
	time.Sleep(panicTimer * 2) // 等待 panic 触发
	a.Equal(srv2.State(), ServiceFailed)

	// 出错后，还能正确运行和结束
	srv2.Run()
	time.Sleep(200 * time.Microsecond) // 等待服务启动完成
	srv2.Stop()

	// srv3

	m.AddService(srv3, "srv3")
	srv3 := m.services[2]
	srv3.Run()
	time.Sleep(200 * time.Microsecond) // 等待服务启动完成
	a.Equal(srv3.State(), ServiceRunning)
	time.Sleep(500 * time.Microsecond) // 等待超过返回错误
	a.Equal(srv3.State(), ServiceFailed)
	a.NotNil(srv3.Err())

	// 再次运行
	srv3.Run()
	time.Sleep(200 * time.Microsecond) // 等待服务启动完成
	a.Equal(srv3.State(), ServiceRunning)
	srv3.Stop()
	a.Equal(srv3.State(), ServiceStop)
}

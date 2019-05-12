// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package module

import (
	"errors"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/issue9/assert"

	"github.com/issue9/web/internal/webconfig"
)

func TestNew(t *testing.T) {
	a := assert.New(t)

	ms, err := NewModules(&webconfig.WebConfig{}, logsDefault)
	a.NotError(err).NotNil(ms)
	a.Equal(len(ms.Modules()), 1).
		Equal(ms.modules[0].Name, CoreModuleName).
		NotNil(ms.Mux()).
		Equal(ms.Mux(), ms.router.Mux())
}

func TestModules_Init(t *testing.T) {
	a := assert.New(t)
	ms, err := NewModules(&webconfig.WebConfig{}, logsDefault)
	a.NotError(err).NotNil(ms)

	m1 := ms.NewModule("users1", "user1 module", "users2", "users3")
	srv1, start1, exit1 := buildSrv1()
	m1.AddService(srv1, "服务 1")
	srv2, start2, exit2 := buildSrv1()
	m1.AddService(srv2, "服务 2")
	a.Equal(len(ms.services), 0)
	m1.NewTag("v1").
		AddInit(func() error { return errors.New("falid message") }, "安装数据表 users1")

	m2 := ms.NewModule("users2", "user2 module", "users3")
	srv3, start3, exit3 := buildSrv1()
	m2.AddService(srv3, "服务 3")
	m2.NewTag("v1").AddInit(func() error { return nil }, "安装数据表 users2")
	a.Equal(len(ms.services), 0)

	m3 := ms.NewModule("users3", "user3 mdoule")
	tag := m3.NewTag("v1")
	tag.AddInit(func() error { return nil }, "安装数据表 users3-1")
	tag.AddInit(func() error { return nil }, "安装数据表 users3-2")
	tag.AddInit(func() error { return nil }, "安装数据表 users3-3")
	a.Equal(len(tag.inits), 3)
	a.Equal(len(ms.services), 0)

	a.Error(ms.Init("v1", nil))            // 出错后中断
	time.Sleep(5 * tickTimer)              // 等待 ms.Init 中的服务结束
	a.NotError(ms.Init("not exists", nil)) // 不存在
	time.Sleep(5 * tickTimer)              // 等待 ms.Init 中的服务结束

	a.Equal(0, len(ms.Services()))
	a.NotError(ms.Init("", log.New(os.Stdout, "", 0)))
	a.Equal(4, len(ms.Services())) // 自带 ms.scheduled

	// 等待启动完成
	wg := &sync.WaitGroup{}
	wg.Add(3)
	go func() {
		<-start1
		wg.Done()
	}()
	go func() {
		<-start2
		wg.Done()
	}()
	go func() {
		<-start3
		wg.Done()
	}()
	wg.Wait()

	ms.Stop()

	// 等待停止
	wg = &sync.WaitGroup{}
	wg.Add(3)
	go func() {
		<-exit1
		wg.Done()
	}()
	go func() {
		<-exit2
		wg.Done()
	}()
	go func() {
		<-exit3
		wg.Done()
	}()
	wg.Wait()
	for _, srv := range ms.Services() {
		a.Equal(srv.State(), ServiceStop)
	}
}

func TestModules_Tags(t *testing.T) {
	a := assert.New(t)
	ms, err := NewModules(&webconfig.WebConfig{}, logsDefault)
	a.NotError(err).NotNil(ms)

	m1 := ms.NewModule("users1", "user1 module", "users2", "users3")
	m1.NewTag("v1").
		AddInit(func() error { return errors.New("falid message") }, "安装数据表 users1")
	m1.NewTag("v2")

	m2 := ms.NewModule("users2", "user2 module", "users3")
	m2.NewTag("v1").AddInit(func() error { return nil }, "安装数据表 users2")
	m2.NewTag("v3")

	m3 := ms.NewModule("users3", "user3 mdoule")
	tag := m3.NewTag("v1")
	tag.AddInit(func() error { return nil }, "安装数据表 users3-1")
	tag.AddInit(func() error { return nil }, "安装数据表 users3-2")
	m3.NewTag("v4")

	tags := ms.Tags()
	a.Equal(tags, []string{"v1", "v2", "v3", "v4"})
}

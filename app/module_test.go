// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"errors"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/issue9/assert"
)

func TestModule_NewTag(t *testing.T) {
	a := assert.New(t)
	app := newApp(a)
	m := newModule(app, "user1", "user1 desc")
	a.NotNil(m)

	v := m.NewTag("0.1.0")
	a.NotNil(v).NotNil(m.tags["0.1.0"])
	v.AddInit(nil, "title1")
	a.Equal(v.inits[0].title, "title1")

	vv := m.NewTag("0.1.0")
	a.Equal(vv, v)

	v2 := m.NewTag("0.2.0")
	a.NotEqual(v2, v)
}

func TestModule_Plugin(t *testing.T) {
	a := assert.New(t)

	app := newApp(a)
	m := newModule(app, "user1", "user1 desc")
	a.NotNil(m)

	a.Panic(func() {
		m.Plugin("p1", "p1 desc")
	})

	m = newModule(app, "", "")
	a.NotPanic(func() {
		m.Plugin("p1", "p1 desc")
	})
}

func TestModule_AddInit(t *testing.T) {
	a := assert.New(t)
	app := newApp(a)
	m := newModule(app, "m1", "m1 desc")
	a.NotNil(m)

	a.Nil(m.inits)
	m.AddInit(func() error { return nil }, "t1")
	a.Equal(len(m.inits), 1).
		Equal(m.inits[0].title, "t1").
		NotNil(m.inits[0].f)

	m.AddInit(func() error { return nil }, "t1")
	a.Equal(len(m.inits), 2).
		Equal(m.inits[1].title, "t1").
		NotNil(m.inits[1].f)

	m.AddInit(func() error { return nil }, "t1")
	a.Equal(len(m.inits), 3).
		Equal(m.inits[2].title, "t1").
		NotNil(m.inits[2].f)
}

func TestApp_Init(t *testing.T) {
	a := assert.New(t)
	app := newApp(a)

	m1 := app.NewModule("users1", "user1 module", "users2", "users3")
	srv1, start1, exit1 := buildSrv1()
	m1.AddService(srv1, "服务 1")
	srv2, start2, exit2 := buildSrv1()
	m1.AddService(srv2, "服务 2")
	a.Equal(len(app.services), 0)
	m1.NewTag("v1").
		AddInit(func() error { return errors.New("falid message") }, "安装数据表 users1")

	m2 := app.NewModule("users2", "user2 module", "users3")
	srv3, start3, exit3 := buildSrv1()
	m2.AddService(srv3, "服务 3")
	m2.NewTag("v1").AddInit(func() error { return nil }, "安装数据表 users2")
	a.Equal(len(app.services), 0)

	m3 := app.NewModule("users3", "user3 mdoule")
	tag := m3.NewTag("v1")
	tag.AddInit(func() error { return nil }, "安装数据表 users3-1")
	tag.AddInit(func() error { return nil }, "安装数据表 users3-2")
	tag.AddInit(func() error { return nil }, "安装数据表 users3-3")
	a.Equal(len(tag.inits), 3)
	a.Equal(len(app.services), 0)

	a.Error(app.Init("v1", nil))            // 出错后中断
	time.Sleep(5 * tickTimer)               // 等待 app.Init 中的服务结束
	a.NotError(app.Init("not exists", nil)) // 不存在
	time.Sleep(5 * tickTimer)               // 等待 app.Init 中的服务结束

	a.Equal(0, len(app.Services()))
	a.NotError(app.Init("", log.New(os.Stdout, "", 0)))
	a.Equal(4, len(app.Services())) // 自带 app.scheduled

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

	app.Stop()

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
	for _, srv := range app.Services() {
		a.Equal(srv.State(), ServiceStop)
	}
}

func TestApp_Tags(t *testing.T) {
	a := assert.New(t)
	app := newApp(a)

	m1 := app.NewModule("users1", "user1 module", "users2", "users3")
	m1.NewTag("v1").
		AddInit(func() error { return errors.New("falid message") }, "安装数据表 users1")
	m1.NewTag("v2")

	m2 := app.NewModule("users2", "user2 module", "users3")
	m2.NewTag("v1").AddInit(func() error { return nil }, "安装数据表 users2")
	m2.NewTag("v3")

	m3 := app.NewModule("users3", "user3 mdoule")
	tag := m3.NewTag("v1")
	tag.AddInit(func() error { return nil }, "安装数据表 users3-1")
	tag.AddInit(func() error { return nil }, "安装数据表 users3-2")
	m3.NewTag("v4")

	tags := app.Tags()
	a.Equal(tags, []string{"v1", "v2", "v3", "v4"})
}

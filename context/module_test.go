// SPDX-License-Identifier: MIT

package context

import (
	"errors"
	"log"
	"os"
	"testing"
	"time"
	"unicode"

	"github.com/issue9/assert"
)

func job(time.Time) error {
	println("job")
	return nil
}

func TestModuleInitFuncName(t *testing.T) {
	a := assert.New(t)

	a.True(unicode.IsUpper(rune(moduleInstallFuncName[0])))
}

func TestModule_NewTag(t *testing.T) {
	a := assert.New(t)
	srv := newServer(a)
	m := srv.NewModule("user1", "user1 desc")
	a.NotNil(m)

	v := m.NewTag("0.1.0")
	a.NotNil(v).NotNil(m.tags["0.1.0"])
	v.AddInit(nil, "title1")
	a.Equal(v.inits[0].title, "title1")

	vv := m.NewTag("0.1.0")
	a.Equal(vv, v).Equal(vv.m, m)

	v2 := m.NewTag("0.2.0")
	a.NotEqual(v2, v).Equal(v2.m, m)
}

func TestModule_AddInit(t *testing.T) {
	a := assert.New(t)
	srv := newServer(a)

	m := srv.NewModule("m1", "m1 desc")
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

	srv.initDeps("", srv.logs.INFO())
	a.Panic(func() {
		m.AddInit(func() error { return nil }, "t1")
	})
}

func TestWeb_Tags(t *testing.T) {
	a := assert.New(t)
	srv := newServer(a)

	m1 := srv.NewModule("users1", "user1 module", "users2", "users3")
	m1.NewTag("v1").
		AddInit(func() error { return errors.New("failed message") }, "安装数据表 users1")
	m1.NewTag("v2")

	m2 := srv.NewModule("users2", "user2 module", "users3")
	m2.NewTag("v1").AddInit(func() error { return nil }, "安装数据表 users2")
	m2.NewTag("v3")

	m3 := srv.NewModule("users3", "user3 module")
	tag := m3.NewTag("v1")
	tag.AddInit(func() error { return nil }, "安装数据表 users3-1")
	tag.AddInit(func() error { return nil }, "安装数据表 users3-2")
	m3.NewTag("v4")

	tags := srv.Tags()
	a.Equal(3, len(tags))
	a.Equal(tags["users1"], []string{"v1", "v2"}).
		Equal(tags["users2"], []string{"v1", "v3"}).
		Equal(tags["users3"], []string{"v1", "v4"})
}

func TestWeb_Init(t *testing.T) {
	a := assert.New(t)
	srv := newServer(a)

	m1 := srv.NewModule("m1", "m1 desc", "m2")
	m1.AddCron("test cron", job, "* * 8 * * *", true)
	m1.AddAt("test cron", job, time.Now().Add(-time.Hour), true)

	m2 := srv.NewModule("m2", "m2 desc")
	m2.AddTicker("ticker test", job, 5*time.Second, false, false)

	a.Equal(len(srv.Modules()), 2)

	a.Equal(0, len(srv.Services().Jobs())) // 需要初始化模块之后，才有计划任务
	a.NotError(srv.Init("", log.New(os.Stdout, "[INFO]", 0)))
	a.Equal(3, len(srv.Services().Jobs()))

	// 不能多次调用
	a.Equal(srv.Init("", log.New(os.Stdout, "[INFO]", 0)), ErrInited)
}

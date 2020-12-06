// SPDX-License-Identifier: MIT

package web

import (
	"errors"
	"log"
	"os"
	"testing"
	"time"
	"unicode"

	"github.com/issue9/assert"

	"github.com/issue9/web/internal/dep"
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
	m := NewModule("user1", "user1 desc")
	a.NotNil(m)

	v := m.NewTag("0.1.0")
	a.NotNil(v)
	v.AddInit("title1", nil)
	def, ok := v.(*dep.Module)
	a.True(ok).Equal(def.ID(), "user1") // 与模块相同的 ID

	vv := m.NewTag("0.1.0")
	a.Equal(vv, v)

	v2 := m.NewTag("0.2.0")
	a.NotEqual(v2, v)
}

func TestServer_Tags(t *testing.T) {
	a := assert.New(t)

	m1 := NewModule("users1", "user1 module", "users2", "users3")
	a.NotNil(m1)
	t1 := m1.NewTag("v1")
	a.NotNil(t1)
	t1.AddInit("安装数据表 users1", func() error { return errors.New("failed message") })
	m1.NewTag("v2")

	m2 := NewModule("users2", "user2 module", "users3")
	a.NotNil(m2)
	t2 := m2.NewTag("v1")
	a.NotNil(t2)
	t2.AddInit("安装数据表 users2", func() error { return nil })
	m2.NewTag("v3")

	m3 := NewModule("users3", "user3 module")
	a.NotNil(m3)
	tag := m3.NewTag("v1")
	a.NotNil(tag)
	tag.AddInit("安装数据表 users3-1", func() error { return nil })
	tag.AddInit("安装数据表 users3-2", func() error { return nil })
	m3.NewTag("v4")

	srv := newServer(a)
	a.NotError(srv.AddModule(m1, m2, m3))
	tags := srv.Tags()
	a.Equal(3, len(tags))
	a.Equal(tags["users1"], []string{"v1", "v2"}).
		Equal(tags["users2"], []string{"v1", "v3"}).
		Equal(tags["users3"], []string{"v1", "v4"})
}

func TestServer_initModules(t *testing.T) {
	a := assert.New(t)

	m1 := NewModule("m1", "m1 desc", "m2")
	a.NotNil(m1)
	m1.AddCron("test cron", job, "* * 8 * * *", true)
	m1.AddAt("test cron", job, time.Now().Add(-time.Hour), true)

	m2 := NewModule("m2", "m2 desc")
	a.NotNil(m2)
	m2.AddTicker("ticker test", job, 5*time.Second, false, false)

	srv := newServer(a)
	srv.AddModule(m1, m2)
	a.Equal(len(srv.Modules()), 2)

	a.Equal(0, len(srv.Services().Jobs())) // 需要初始化模块之后，才有计划任务
	a.NotError(srv.initModules(log.New(os.Stdout, "[INFO]", 0)))
	a.Equal(3, len(srv.Services().Jobs()))

	// 不能多次调用
	a.Equal(srv.initModules(log.New(os.Stdout, "[INFO]", 0)), ErrInited)
}

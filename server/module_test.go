// SPDX-License-Identifier: MIT

package server

import (
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"github.com/issue9/assert"
)

func job(time.Time) error {
	println("job")
	return nil
}

func TestModule_NewTag(t *testing.T) {
	a := assert.New(t)
	srv := newServer(a)
	m := srv.newModule("user1", "user1 desc")
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

func TestModules_Init(t *testing.T) {
	a := assert.New(t)
	srv := newServer(a)

	m1 := srv.New("m1", "m1 desc", "m2")
	m1.AddCron("test cron", job, "* * 8 * * *", true)
	m1.AddAt("test cron", job, "2020-01-02 17:55:11", true)

	m2 := srv.New("m2", "m2 desc")
	m2.AddTicker("ticker test", job, 5*time.Second, false, false)

	a.Equal(len(srv.Modules()), 2)

	a.NotError(srv.Init("", log.New(os.Stdout, "[INFO]", 0)))
}

func TestModule_Plugin(t *testing.T) {
	a := assert.New(t)
	srv := newServer(a)

	m := srv.newModule("user1", "user1 desc")
	a.NotNil(m)

	a.Panic(func() {
		m.Plugin("p1", "p1 desc")
	})

	m = srv.newModule("", "")
	a.NotPanic(func() {
		m.Plugin("p1", "p1 desc")
	})
}

func TestModule_AddInit(t *testing.T) {
	a := assert.New(t)
	srv := newServer(a)

	m := srv.newModule("m1", "m1 desc")
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

func TestModules_Tags(t *testing.T) {
	a := assert.New(t)
	srv := newServer(a)

	m1 := srv.New("users1", "user1 module", "users2", "users3")
	m1.NewTag("v1").
		AddInit(func() error { return errors.New("failed message") }, "安装数据表 users1")
	m1.NewTag("v2")

	m2 := srv.New("users2", "user2 module", "users3")
	m2.NewTag("v1").AddInit(func() error { return nil }, "安装数据表 users2")
	m2.NewTag("v3")

	m3 := srv.New("users3", "user3 module")
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

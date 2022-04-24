// SPDX-License-Identifier: MIT

package app

import (
	"bytes"
	"os"
	"testing"

	"github.com/issue9/assert/v2"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web/server"
)

func TestAppOf_Exec(t *testing.T) {
	a := assert.New(t, false)

	bs := new(bytes.Buffer)
	var action string
	aa := &AppOf[empty]{
		Name:    "test",
		Version: "1.0.0",
		Out:     bs,
		Init: func(s *server.Server, user *empty, act string) error {
			action = act
			return nil
		},
	}
	a.NotError(aa.Exec([]string{"app", "-v"}))
	a.Contains(bs.String(), aa.Version)

	bs.Reset()
	a.NotError(aa.Exec([]string{"app", "-f=./testdata", "-a=install"}))
	a.Equal(action, "install")
}

func TestAppOf_sanitize(t *testing.T) {
	a := assert.New(t, false)

	cmd := &AppOf[empty]{}
	a.ErrorString(cmd.sanitize(), "Name")

	cmd = &AppOf[empty]{Name: "app", Version: "1.1.1"}
	a.ErrorString(cmd.sanitize(), "Init")

	cmd = &AppOf[empty]{
		Name:    "app",
		Version: "1.1.1",
		Init:    func(*server.Server, *empty, string) error { return nil },
	}
	a.NotError(cmd.sanitize())

	a.Equal(cmd.Out, os.Stdout)
}

func TestAppOf_initOptions(t *testing.T) {
	a := assert.New(t, false)

	cmd := &AppOf[empty]{
		Name:    "app",
		Version: "1.1.1",
		Init:    func(*server.Server, *empty, string) error { return nil },
		Catalog: catalog.NewBuilder(),
	}
	a.NotError(cmd.sanitize())
	opt, user, err := cmd.initOptions(os.DirFS("./"))
	a.NotError(err).NotNil(opt).Nil(user)
	a.Nil(opt.Catalog). // 不会传递给 opt.Catalog
		Equal(opt.FileSerializers, cmd.FileSerializers).
		False(opt.Catalog == cmd.Catalog) // 不指向同一个对象

	// 包含 Options
	cmd = &AppOf[empty]{
		Name:    "app",
		Version: "1.1.1",
		Init:    func(*server.Server, *empty, string) error { return nil },
		Catalog: catalog.NewBuilder(),
		Options: func(o *server.Options) error {
			o.Catalog = catalog.NewBuilder() // 改变了 Catalog
			return nil
		},
	}
	a.NotError(cmd.sanitize())
	opt, user, err = cmd.initOptions(os.DirFS("./testdata"))
	a.NotError(err).NotNil(opt).Nil(user)
	a.NotNil(opt.Catalog).
		Equal(opt.FileSerializers, cmd.FileSerializers).
		False(opt.Catalog == cmd.Catalog) // 不指向同一个对象

	// 包含 ConfigFilename
	cmd2 := &AppOf[userData]{
		Name:           "app",
		Version:        "1.1.1",
		Init:           func(*server.Server, *userData, string) error { return nil },
		ConfigFilename: "user.xml",
	}
	a.NotError(cmd2.sanitize())
	opt2, user2, err := cmd2.initOptions(os.DirFS("./testdata"))
	a.NotError(err).NotNil(opt2).NotNil(user2)
	a.Nil(opt2.Catalog).
		Equal(opt2.FileSerializers, cmd2.FileSerializers).
		Equal(user2.ID, 1)
}

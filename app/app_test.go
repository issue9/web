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

func TestApp_Exec(t *testing.T) {
	a := assert.New(t, false)

	bs := new(bytes.Buffer)
	var action string
	aa := &App[empty]{
		Name:    "test",
		Version: "1.0.0",
		Out:     bs,
		Init: func(s *server.Server, user *empty, act string) error {
			action = act
			return nil
		},
	}
	aa.Exec([]string{"app", "-v"})
	a.Contains(bs.String(), aa.Version)

	bs.Reset()
	aa.Exec([]string{"app", "-f=./testdata", "-a=install"})
	a.Equal(action, "install")
}

func TestApp_sanitize(t *testing.T) {
	a := assert.New(t, false)

	cmd := &App[empty]{}
	a.ErrorString(cmd.sanitize(), "Name")

	cmd = &App[empty]{Name: "app", Version: "1.1.1"}
	a.ErrorString(cmd.sanitize(), "Init")

	cmd = &App[empty]{
		Name:    "app",
		Version: "1.1.1",
		Init:    func(*server.Server, *empty, string) error { return nil },
	}
	a.NotError(cmd.sanitize())

	a.Equal(cmd.Out, os.Stdout)
}

func TestCommand_initOptions(t *testing.T) {
	a := assert.New(t, false)

	cmd := &App[empty]{
		Name:    "app",
		Version: "1.1.1",
		Init:    func(*server.Server, *empty, string) error { return nil },
		Catalog: catalog.NewBuilder(),
	}
	a.NotError(cmd.sanitize())
	opt, user, err := cmd.initOptions(os.DirFS("./"))
	a.NotError(err).NotNil(opt).Nil(user)
	a.NotNil(opt.Catalog).
		Equal(opt.Files, cmd.Files).
		True(opt.Catalog == cmd.Catalog)

	// 包含 Options
	cmd = &App[empty]{
		Name:    "app",
		Version: "1.1.1",
		Init:    func(*server.Server, *empty, string) error { return nil },
		Catalog: catalog.NewBuilder(),
		Options: func(o *server.Options) { o.Catalog = catalog.NewBuilder() }, // 改变了 Catalog
	}
	a.NotError(cmd.sanitize())
	opt, user, err = cmd.initOptions(os.DirFS("./testdata"))
	a.NotError(err).NotNil(opt).Nil(user)
	a.NotNil(opt.Catalog).
		Equal(opt.Files, cmd.Files).
		False(opt.Catalog == cmd.Catalog) // 不指向同一个对象

	// 包含 ConfigFilename
	cmd2 := &App[userData]{
		Name:           "app",
		Version:        "1.1.1",
		Init:           func(*server.Server, *userData, string) error { return nil },
		ConfigFilename: "user.xml",
	}
	a.NotError(cmd2.sanitize())
	opt2, user2, err := cmd2.initOptions(os.DirFS("./testdata"))
	a.NotError(err).NotNil(opt2).NotNil(user2)
	a.NotNil(opt2.Catalog).
		Equal(opt2.Files, cmd2.Files).
		Equal(user2.ID, 1)
}

// SPDX-License-Identifier: MIT

package app

import (
	"bytes"
	"os"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/web/server"
	"golang.org/x/text/message/catalog"
)

func TestApp_Exec(t *testing.T) {
	a := assert.New(t, false)

	bs := new(bytes.Buffer)
	var action string
	aa := &App{
		Name:    "test",
		Version: "1.0.0",
		Out:     bs,
		Init: func(s *server.Server, act string) error {
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

	cmd := &App{}
	a.ErrorString(cmd.sanitize(), "Name")

	cmd = &App{Name: "app", Version: "1.1.1"}
	a.ErrorString(cmd.sanitize(), "Init")

	cmd = &App{
		Name:    "app",
		Version: "1.1.1",
		Init:    func(*server.Server, string) error { return nil },
	}
	a.NotError(cmd.sanitize())

	a.Equal(cmd.Out, os.Stdout)
}

func TestApp_initOptions(t *testing.T) {
	a := assert.New(t, false)

	cmd := &App{
		Name:    "app",
		Version: "1.1.1",
		Init:    func(*server.Server, string) error { return nil },
		Catalog: catalog.NewBuilder(),
	}
	a.NotError(cmd.sanitize())
	opt, err := cmd.initOptions(os.DirFS("./"))
	a.NotError(err).NotNil(opt)
	a.NotNil(opt.Catalog).
		Equal(opt.Files, cmd.Files).
		True(opt.Catalog == cmd.Catalog)

	// 包含 ConfigFilename
	cmd = &App{
		Name:           "app",
		Version:        "1.1.1",
		Init:           func(*server.Server, string) error { return nil },
		ConfigFilename: "web.yaml",
	}
	a.NotError(cmd.sanitize())
	opt, err = cmd.initOptions(os.DirFS("./testdata"))
	a.NotError(err).NotNil(opt)
	a.NotNil(opt.Catalog).Equal(opt.Files, cmd.Files)

	// 包含 Options
	cmd = &App{
		Name:    "app",
		Version: "1.1.1",
		Init:    func(*server.Server, string) error { return nil },
		Catalog: catalog.NewBuilder(),
		Options: func(o *server.Options) { o.Catalog = catalog.NewBuilder() }, // 改变了 Catalog
	}
	a.NotError(cmd.sanitize())
	opt, err = cmd.initOptions(os.DirFS("./config"))
	a.NotError(err).NotNil(opt)
	a.NotNil(opt.Catalog).
		Equal(opt.Files, cmd.Files).
		False(opt.Catalog == cmd.Catalog) // 不指向同一个对象
}

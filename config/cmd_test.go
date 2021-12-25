// SPDX-License-Identifier: MIT

package config

import (
	"os"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/web/server"
	"golang.org/x/text/message/catalog"
)

func TestCommand_sanitize(t *testing.T) {
	a := assert.New(t, false)

	cmd := &Command{}
	a.ErrorString(cmd.sanitize(), "Name")

	cmd = &Command{Name: "app", Version: "1.1.1"}
	a.ErrorString(cmd.sanitize(), "Init")

	cmd = &Command{
		Name:    "app",
		Version: "1.1.1",
		Init:    func(*server.Server, string) error { return nil },
	}
	a.NotError(cmd.sanitize())

	a.Equal(cmd.Out, os.Stdout)
}

func TestCommand_initOptions(t *testing.T) {
	a := assert.New(t, false)

	cmd := &Command{
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
	cmd = &Command{
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
	cmd = &Command{
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

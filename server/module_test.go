// SPDX-License-Identifier: MIT

package server

import (
	"io/fs"
	"os"
	"testing"

	"github.com/issue9/assert/v2"

	"github.com/issue9/web/internal/filesystem"
)

var _ fs.GlobFS = &Module{}

func TestServer_NewModule(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, &Options{FS: os.DirFS("./")})

	m := srv.NewModule("testdata")
	a.NotNil(m).
		Equal(m.ID(), "testdata").
		Equal(m.Server(), srv)

	bs, err := fs.ReadFile(m, "file1.txt")
	a.NotError(err).Equal(bs, []byte("file1"))

	a.PanicString(func() {
		srv.NewModule("testdata")
	}, "存在同名模块")

	a.PanicString(func() {
		srv.NewModule("//")
	}, "无效的 id 格式")

	a.PanicString(func() {
		srv.NewModule("")
	}, "无效的 id 格式")
}

func TestModule_BuildID(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, nil)

	m := srv.NewModule("id")
	a.Equal(m.ID(), "id").Equal(m.BuildID("1"), "id_1")
}

func TestModule_Glob(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, &Options{FS: os.DirFS("./")})

	m := srv.NewModule("testdata")
	a.True(filesystem.ExistsFS(m, "file1.txt"))
	a.False(filesystem.ExistsFS(m, "not-exists.txt"))
	a.False(filesystem.ExistsFS(m, "servertest.go"))
	matches, err := fs.Glob(m, "*.go")
	a.NotError(err).Equal(matches, []string{"result.pb.go"})

	matches, err = fs.Glob(m, "*.pem")
	a.NotError(err).Equal(matches, []string{"cert.pem", "key.pem"})

	// AddFS
	m.AddFS(os.DirFS("./servertest"))
	a.True(filesystem.ExistsFS(m, "servertest.go"))
	matches, err = fs.Glob(m, "*.go")
	a.NotError(err).Equal(matches, []string{"result.pb.go"})
	matches, err = fs.Glob(m, "*.pem")
	a.NotError(err).Equal(matches, []string{"cert.pem", "key.pem"})

	// 反向顺序

	m = srv.NewModule("servertest")
	a.False(filesystem.ExistsFS(m, "file1.txt"))
	a.False(filesystem.ExistsFS(m, "not-exists.txt"))
	a.True(filesystem.ExistsFS(m, "servertest.go"))
	matches, err = fs.Glob(m, "*.go")
	a.NotError(err).Equal(matches, []string{"server.go", "servertest.go"})
	matches, err = fs.Glob(m, "*.pem")
	a.NotError(err).Empty(matches)

	// AddFS
	m.AddFS(os.DirFS("./testdata"))
	a.True(filesystem.ExistsFS(m, "file1.txt"))
	matches, err = fs.Glob(m, "*.go")
	a.NotError(err).Equal(matches, []string{"server.go", "servertest.go"})
	matches, err = fs.Glob(m, "*.pem")
	a.NotError(err).Equal(matches, []string{"cert.pem", "key.pem"})
}

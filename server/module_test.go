// SPDX-License-Identifier: MIT

package server

import (
	"errors"
	"io/fs"
	"os"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/localeutil"
	"golang.org/x/text/language"
)

var _ fs.GlobFS = &Module{}

func TestServer_NewModule(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, &Options{FS: os.DirFS("./")})
	p := srv.NewPrinter(language.SimplifiedChinese)

	desc := localeutil.Phrase("lang")
	m := srv.NewModule("testdata", desc)
	a.NotNil(m).
		Equal(m.ID(), "testdata").
		Equal(m.Server(), srv).
		Equal(m.Description().LocaleString(p), "hans")

	bs, err := fs.ReadFile(m, "file1.txt")
	a.NotError(err).Equal(bs, []byte("file1"))

	a.PanicString(func() {
		srv.NewModule("testdata", desc)
	}, "存在同名模块")

	a.PanicString(func() {
		srv.NewModule("//", desc)
	}, "无效的 id 格式")

	a.PanicString(func() {
		srv.NewModule("", desc)
	}, "无效的 id 格式")
}

func TestServer_Modules(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, &Options{FS: os.DirFS("./")})

	srv.NewModule("m1", localeutil.Phrase("lang"))
	srv.NewModule("m2", localeutil.Phrase("m2 desc"))
	srv.NewModule("m3", localeutil.Phrase("m3 desc"))

	p := srv.NewPrinter(language.SimplifiedChinese)
	mods := srv.Modules(p)
	a.Equal(mods, map[string]string{
		"m1": "hans",
		"m2": "m2 desc",
		"m3": "m3 desc",
	})
}

func TestModule_BuildID(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, nil)

	m := srv.NewModule("id", localeutil.Phrase("lang"))
	a.Equal(m.ID(), "id").Equal(m.BuildID("_1"), "id_1")
}

func TestModule_Glob(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, &Options{FS: os.DirFS("./")})

	existsFS := func(fsys fs.FS, p string) bool {
		_, err := fs.Stat(fsys, p)
		return err == nil || errors.Is(err, fs.ErrExist)
	}

	m := srv.NewModule("testdata", localeutil.Phrase("lang"))
	a.True(existsFS(m, "file1.txt"))
	a.False(existsFS(m, "not-exists.txt"))
	a.False(existsFS(m, "servertest.go"))
	matches, err := fs.Glob(m, "*.cnf")
	a.NotError(err).Equal(matches, []string{"req.cnf"})

	matches, err = fs.Glob(m, "*.pem")
	a.NotError(err).Equal(matches, []string{"cert.pem", "key.pem"})

	// AddFS
	m.AddFS(os.DirFS("./servertest"))
	a.True(existsFS(m, "servertest.go"))
	matches, err = fs.Glob(m, "*.cnf")
	a.NotError(err).Equal(matches, []string{"req.cnf"})
	matches, err = fs.Glob(m, "*.pem")
	a.NotError(err).Equal(matches, []string{"cert.pem", "key.pem"})

	// 反向顺序

	m = srv.NewModule("servertest", localeutil.Phrase("lang"))
	a.False(existsFS(m, "file1.txt"))
	a.False(existsFS(m, "not-exists.txt"))
	a.True(existsFS(m, "servertest.go"))
	matches, err = fs.Glob(m, "*.go")
	a.NotError(err).Equal(matches, []string{"server.go", "servertest.go"})
	matches, err = fs.Glob(m, "*.pem")
	a.NotError(err).Empty(matches)

	// AddFS
	m.AddFS(os.DirFS("./testdata"))
	a.True(existsFS(m, "file1.txt"))
	matches, err = fs.Glob(m, "*.go")
	a.NotError(err).Equal(matches, []string{"server.go", "servertest.go"})
	matches, err = fs.Glob(m, "*.pem")
	a.NotError(err).Equal(matches, []string{"cert.pem", "key.pem"})
}

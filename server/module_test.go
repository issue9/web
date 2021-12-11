// SPDX-License-Identifier: MIT

package server

import (
	"io/fs"
	"os"
	"testing"

	"github.com/issue9/assert/v2"
)

var _ fs.FS = &Module{}

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
}

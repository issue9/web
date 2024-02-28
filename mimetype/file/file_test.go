// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package file

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/web"
	"golang.org/x/text/language"

	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"
)

func newServer(a *assert.Assertion) web.Server {
	o := &server.Options{
		HTTPServer: &http.Server{Addr: ":8080"},
		Language:   language.English,
		Mimetypes: []*server.Mimetype{
			{
				Name:      "application/json",
				Marshal:   func(*web.Context, any) ([]byte, error) { return nil, nil },
				Unmarshal: func(io.Reader, any) error { return nil },
				Problem:   "application/problem+json",
			},
		},
		Logs: &server.Logs{
			Handler:  server.NewTermHandler(os.Stderr, nil),
			Location: true,
			Created:  server.NanoLayout,
			Levels:   server.AllLevels(),
		},
	}
	srv, err := server.New("test", "1.0.0", o)
	a.NotError(err).NotNil(srv)

	return srv
}

func TestServeFileHandler(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a)
	router := srv.Routers().New("def", nil)

	defer servertest.Run(a, srv)()
	defer srv.Close(0)

	a.PanicString(func() {
		ServeFileHandler(nil, "path", "index.html")
	}, "参数 fsys 不能为空")

	a.PanicString(func() {
		ServeFileHandler(os.DirFS("./testdata"), "", "index.html")
	}, "参数 name 不能为空")

	router.Get("/serve/{path}", ServeFileHandler(os.DirFS("./testdata"), "path", "index.html"))
	servertest.Get(a, "http://localhost:8080/serve/file1.txt"). // file1.txt
									Do(nil).
									Status(http.StatusOK).
									StringBody("file1")
	servertest.Get(a, "http://localhost:8080/serve/"). // index.html
								Do(nil).
								Status(http.StatusOK).
								BodyFunc(func(a *assert.Assertion, body []byte) {
			a.True(bytes.HasPrefix(body, []byte("<!DOCTYPE html>")))
		})
}

func TestAttachmentHandler(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a)
	router := srv.Routers().New("def", nil)

	defer servertest.Run(a, srv)()
	defer srv.Close(0)

	a.PanicString(func() {
		AttachmentHandler(nil, "path", "filename", true)
	}, "参数 fsys 不能为空")

	a.PanicString(func() {
		AttachmentHandler(os.DirFS("./testdata"), "", "filename", true)
	}, "参数 name 不能为空")

	router.Get("/attach/{path}", AttachmentHandler(os.DirFS("./testdata"), "path", "中文", true))

	servertest.Get(a, "http://localhost:8080/attach/file1.txt"). // file1.txt
									Do(nil).
									Status(http.StatusOK).
									Header(header.ContentDisposition, "inline; filename="+url.QueryEscape("中文"))
}

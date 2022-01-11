// SPDX-License-Identifier: MIT

// Package servertest 针对 server 的测试用例
package servertest

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"os"
	"time"

	"github.com/issue9/assert/v2"
	"github.com/issue9/localeutil"
	"github.com/issue9/logs/v3"
	"golang.org/x/text/language"

	"github.com/issue9/web/serialization/gob"
	"github.com/issue9/web/serialization/text"
	"github.com/issue9/web/server"
)

type Server struct {
	a    *assert.Assertion
	s    *server.Server
	exit chan struct{}
}

// NewServer 声明一个 server 实例
func NewServer(a *assert.Assertion, o *server.Options) *Server {
	if o == nil {
		o = &server.Options{Port: ":8080"}
	}
	if o.Logs == nil { // 默认重定向到 os.Stderr
		l, err := logs.New(nil)
		a.NotError(err).NotNil(l)

		a.NotError(l.SetOutput(logs.LevelDebug, os.Stderr))
		a.NotError(l.SetOutput(logs.LevelError, os.Stderr))
		a.NotError(l.SetOutput(logs.LevelCritical, os.Stderr))
		a.NotError(l.SetOutput(logs.LevelInfo, os.Stdout))
		a.NotError(l.SetOutput(logs.LevelTrace, os.Stdout))
		a.NotError(l.SetOutput(logs.LevelWarn, os.Stdout))
		o.Logs = l
	}

	srv, err := server.New("app", "0.1.0", o)
	a.NotError(err).NotNil(srv)
	a.Equal(srv.Name(), "app").Equal(srv.Version(), "0.1.0")

	// locale
	b := srv.Locale().Builder()
	a.NotError(b.SetString(language.Und, "lang", "und"))
	a.NotError(b.SetString(language.SimplifiedChinese, "lang", "hans"))
	a.NotError(b.SetString(language.TraditionalChinese, "lang", "hant"))

	// mimetype
	a.NotError(srv.Mimetypes().Add(json.Marshal, json.Unmarshal, "application/json"))
	a.NotError(srv.Mimetypes().Add(xml.Marshal, xml.Unmarshal, "application/xml"))
	a.NotError(srv.Mimetypes().Add(gob.Marshal, gob.Unmarshal, server.DefaultMimetype))
	a.NotError(srv.Mimetypes().Add(text.Marshal, text.Unmarshal, text.Mimetype))

	srv.AddResult(411, "41110", localeutil.Phrase("41110"))

	return &Server{
		s:    srv,
		a:    a,
		exit: make(chan struct{}, 1),
	}
}

func (s *Server) Server() *server.Server { return s.s }

func (s *Server) GoServe() {
	go func() {
		err := s.s.Serve()
		s.a.Error(err).ErrorIs(err, http.ErrServerClosed, "错误信息为:%v", err)
		s.exit <- struct{}{}
	}()
	// 等待 srv.Serve() 启动完毕，不同机器可能需要的时间会不同
	time.Sleep(5000 * time.Microsecond)
}

func (s *Server) Wait() { <-s.exit }

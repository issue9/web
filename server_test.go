// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/issue9/assert/v4"
	"github.com/issue9/cache/caches/memory"
	"github.com/issue9/logs/v7"
	"github.com/issue9/unique/v2"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/internal/locale"
)

var _ Locale = &locale.Locale{}

type testServer struct {
	*InternalServer
	logBuf *bytes.Buffer
}

// 所有输出报头都是 201
func onContext(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request) {
	w.WriteHeader(http.StatusAccepted)
	return w, r
}

func newTestServer(a *assert.Assertion) *testServer {
	c := catalog.NewBuilder()
	a.NotError(c.SetString(language.SimplifiedChinese, "lang", "cn")).
		NotError(c.SetString(language.TraditionalChinese, "lang", "tw"))

	l := language.SimplifiedChinese
	p := message.NewPrinter(l, message.Catalog(c))

	logBuf := new(bytes.Buffer)
	log := logs.New(
		logs.NewTextHandler(logBuf, os.Stderr),
		logs.WithCreated(logs.NanoLayout),
		logs.WithLocation(true),
		logs.WithLevels(logs.AllLevels()...),
		logs.WithLocale(p),
	)
	a.NotNil(log)

	srv := &testServer{logBuf: logBuf}

	cc, gc := memory.New()
	u := unique.NewNumber(100)
	srv.InternalServer = InternalNewServer(srv, "test", "1.0.0", time.Local, log, u.String, locale.New(l, nil, c), cc, newCodec(a), header.RequestIDKey, "")
	srv.Services().Add(Phrase("unique"), u)
	srv.Services().AddTicker(Phrase("gc memory"), func(t time.Time) error { gc(t); return nil }, time.Minute, false, false)

	srv.Problems().Add(411, &LocaleProblem{ID: "41110", Title: Phrase("41110 title"), Detail: Phrase("41110 detail")})

	return srv
}

func (s *testServer) Close(time.Duration) { s.InternalServer.Close() }

func (s *testServer) Serve() error { panic("未实现") }

func (s *testServer) State() State { panic("未实现") }

func TestOnContextFunc(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a)
	s.OnContext(onContext)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/path", nil)
	ctx := s.NewContext(w, r, nil)
	ctx.WriteHeader(http.StatusForbidden)
	a.Equal(ctx.status, http.StatusForbidden).
		Equal(w.Result().StatusCode, http.StatusAccepted)
}

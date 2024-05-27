// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/issue9/assert/v4"
	"github.com/issue9/cache/caches/memory"
	"github.com/issue9/logs/v7"
	"github.com/issue9/mux/v9/header"
	"github.com/issue9/unique/v2"
	"golang.org/x/text/language"

	"github.com/issue9/web/internal/locale"
)

var _ Locale = &locale.Locale{}

type testServer struct {
	*InternalServer
	logBuf *bytes.Buffer
}

func onExitContext(ctx *Context, status int) {
	ctx.Header().Set("exit-status", strconv.Itoa(status))
}

func newTestServer(a *assert.Assertion) *testServer {
	l := locale.New(language.SimplifiedChinese, nil)
	a.NotError(l.SetString(language.SimplifiedChinese, "lang", "cn")).
		NotError(l.SetString(language.TraditionalChinese, "lang", "tw"))

	logBuf := new(bytes.Buffer)
	log := logs.New(
		logs.NewTextHandler(logBuf, os.Stderr),
		logs.WithCreated(logs.NanoLayout),
		logs.WithLocation(true),
		logs.WithLevels(logs.AllLevels()...),
		logs.WithLocale(l.Printer()),
	)
	a.NotNil(log)

	srv := &testServer{logBuf: logBuf}

	cc := memory.New()
	u := unique.NewNumber(100)
	srv.InternalServer = InternalNewServer(srv, "test", "1.0.0", time.Local, log, u.String, l, cc, newCodec(a), header.XRequestID, "", nil)
	srv.Services().Add(Phrase("unique"), u)

	srv.Problems().Add(411, &LocaleProblem{ID: "41110", Title: Phrase("41110 title"), Detail: Phrase("41110 detail")})

	return srv
}

func (s *testServer) Close(time.Duration) { s.InternalServer.Close() }

func (s *testServer) Serve() error { panic("未实现") }

func (s *testServer) State() State { panic("未实现") }

func TestOnExitContextFunc(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a)
	s.OnExitContext(onExitContext)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/path", nil)
	ctx := s.NewContext(w, r, nil)
	ctx.WriteHeader(http.StatusAccepted)
	s.freeContext(ctx)
	a.Equal(ctx.Header().Get("exit-status"), "202")
}

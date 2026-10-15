// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/issue9/assert/v5"
	"github.com/issue9/cache/drivers/memory"
	"github.com/issue9/config"
	"github.com/issue9/logs/v7"
	"github.com/issue9/mux/v10/header"
	"github.com/issue9/unique/v2"
	"golang.org/x/text/language"
)

var _ context.Context = &Server{}

type testServer struct {
	*Server
	logBuf *bytes.Buffer
}

// 创建一个自定义了日志系统的服务
func newTestServer(a *assert.Assertion) *testServer {
	logBuf := new(bytes.Buffer)
	log := logs.New(
		logs.NewTextHandler(logBuf, os.Stderr),
		logs.WithCreated(logs.NanoLayout),
		logs.WithLocation(true),
		logs.WithLevels(logs.AllLevels()...),
	)
	a.NotNil(log)

	conf := config.Dir(nil, "./")

	srv := &testServer{logBuf: logBuf}

	cc := memory.New()
	u := unique.NewNumber(100)
	startup := NewHTTPServerStartup(&http.Server{Addr: ":8080"})
	srv.Server = NewServer("test", "1.0.0", conf, startup, time.Local, language.SimplifiedChinese, log, u.String, cc, newCodec(a), header.XRequestID, "")
	srv.Services().Add(Phrase("unique"), u)

	a.NotError(srv.Locale().SetString(language.SimplifiedChinese, "lang", "cn")).
		NotError(srv.Locale().SetString(language.TraditionalChinese, "lang", "tw"))

	srv.Problems().Add(411, &LocaleProblem{ID: "41110", Title: Phrase("41110 title"), Detail: Phrase("41110 detail")})

	return srv
}

func onExitContext(ctx *Context, status int) {
	ctx.Header().Set("exit-status", strconv.Itoa(status))
}

func TestOnExitContextFunc(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a)
	s.OnExitContext(onExitContext)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/path", nil)
	ctx := s.NewContext(w, r, nil)
	ctx.WriteHeader(http.StatusAccepted)
	ctx.freeContext()
	a.Equal(ctx.Header().Get("exit-status"), "202")
}

// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package startuptest

import (
	"crypto/tls"
	"net/http"

	"github.com/issue9/assert/v6"
	"github.com/issue9/assert/v6/rest"
	"github.com/issue9/web/internal/servertest"

	"github.com/issue9/web"
)

type httpStartup struct {
	web.Startup
	st *servertest.Server
}

// NewHTTPServerStartup 创建一个基于 [http.Server] 的 [web.Startup] 对象
func NewHTTPServerStartup(a *assert.Assertion, sec bool) Startup {
	st := servertest.NewServer(a, sec)

	hs := &http.Server{Addr: st.Port}
	if sec {
		hs.TLSConfig = &tls.Config{
			Certificates: st.Certificates,
		}
	}

	return &httpStartup{
		Startup: web.NewHTTPServerStartup(hs),
		st:      st,
	}
}

func (s *httpStartup) URL() string { return s.st.URL }

func (s *httpStartup) Assertion() *assert.Assertion { return s.st.A }

func (s *httpStartup) Client() *http.Client { return s.st.Client }

func (s *httpStartup) NewRequest(method, path string) *rest.Request {
	return rest.NewRequest(s.Assertion(), method, s.URL()+path).Client(s.st.Client)
}

func (s *httpStartup) Get(path string) *rest.Request { return s.NewRequest(http.MethodGet, path) }

func (s *httpStartup) Put(path string) *rest.Request { return s.NewRequest(http.MethodPut, path) }

func (s *httpStartup) Post(path string) *rest.Request { return s.NewRequest(http.MethodPost, path) }

func (s *httpStartup) Patch(path string) *rest.Request { return s.NewRequest(http.MethodPatch, path) }

func (s *httpStartup) Delete(path string) *rest.Request { return s.NewRequest(http.MethodDelete, path) }

func Run(a *assert.Assertion, s *web.Server) func() { return servertest.Run(a, s) }

// SPDX-License-Identifier: MIT

package registry

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/issue9/assert/v3"
	"github.com/issue9/cache/caches/memory"
	"github.com/issue9/logs/v7"
	"github.com/issue9/unique/v2"
	"golang.org/x/text/language"

	"github.com/issue9/web"
	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/internal/locale"
	"github.com/issue9/web/selector"
)

var _ Registry = &cacheRegistry{}

type testServer struct {
	*web.InternalServer
	httpServer *http.Server
	a          *assert.Assertion
}

func newTestServer(a *assert.Assertion) *testServer {
	log := logs.New(
		logs.NewTextHandler(os.Stderr),
		logs.WithCreated(logs.NanoLayout),
		logs.WithLocation(true),
		logs.WithLevels(logs.AllLevels()...),
	)
	a.NotNil(log)

	srv := &testServer{
		a: a,
		httpServer: &http.Server{
			Addr: ":8080",
		},
	}
	cc, gc := memory.New()
	u := unique.NewNumber(100)
	srv.InternalServer = web.InternalNewServer(srv, "test", "1.0.0", time.Local, log, u.String, locale.New(language.SimplifiedChinese, nil, nil), cc, web.NewCodec(), header.RequestIDKey, "")
	srv.Services().Add(web.Phrase("unique"), u)
	srv.Services().AddTicker(web.Phrase("gc memory"), func(t time.Time) error { gc(t); return nil }, time.Minute, false, false)

	return srv
}

func (s *testServer) Close(shutdownTimeout time.Duration) {
	s.InternalServer.Close()
	s.a.NotError(s.httpServer.Close())
}

func (s *testServer) Serve() error { return s.httpServer.ListenAndServe() }

func (s *testServer) State() web.State { return web.Running }

func TestMarshalPeers(t *testing.T) {
	a := assert.New(t, false)

	peers := []selector.Peer{
		selector.NewPeer("http://localhost:8080"),
		selector.NewPeer("http://localhost:8081"),
	}
	data, err := marshalPeers(peers)
	a.NotError(err).NotEmpty(data)

	peers2, err := unmarshalPeers(func() selector.Peer { return selector.NewPeer("") }, data)
	a.NotError(err).NotNil(peers2).Equal(peers2, peers)
}

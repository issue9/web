// SPDX-License-Identifier: MIT

package web

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"slices"
	"sync"
	"time"

	"github.com/issue9/assert/v3"
	"github.com/issue9/cache"
	"github.com/issue9/cache/caches/memory"
	"github.com/issue9/config"
	"github.com/issue9/logs/v7"
	"github.com/issue9/mux/v7/group"
	"github.com/issue9/mux/v7/types"
	"github.com/issue9/unique/v2"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/internal/locale"
	"github.com/issue9/web/selector"
)

var _ Locale = &locale.Locale{}

type testServer struct {
	a               *assert.Assertion
	logs            *logs.Logs
	logBuf          *bytes.Buffer
	unique          *unique.Unique
	cache           cache.Driver
	routers         *group.GroupOf[HandlerFunc]
	b               *InternalServer
	disableCompress bool
	locale          *locale.Locale
	closes          []func() error
}

func newTestServer(a *assert.Assertion) *testServer {
	c := catalog.NewBuilder()
	a.NotError(c.SetString(language.SimplifiedChinese, "lang", "cn"))
	a.NotError(c.SetString(language.TraditionalChinese, "lang", "tw"))

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

	cc, gc := memory.New()
	srv := &testServer{
		a:      a,
		logs:   log,
		logBuf: logBuf,
		unique: unique.NewNumber(100),
		cache:  cc,
		locale: locale.New(l, nil, c),
		closes: make([]func() error, 0, 10),
	}
	srv.b = InternalNewServer(srv, newCodec(a), header.RequestIDKey, "")
	srv.Services().Add(Phrase("unique"), srv.unique)
	srv.Services().AddTicker(Phrase("gc memory"), func(t time.Time) error { gc(t); return nil }, time.Minute, false, false)

	srv.Problems().Add(411, &LocaleProblem{ID: "41110", Title: Phrase("41110 title"), Detail: Phrase("41110 detail")})

	return srv
}

func (s *testServer) Cache() cache.Cleanable { return s.cache }

func (s *testServer) Close(shutdownTimeout time.Duration) {
	slices.Reverse(s.closes)
	for _, c := range s.closes {
		if err := c(); err != nil {
			fmt.Println(err)
		}
	}
}

func (s *testServer) CompressIsDisable() bool { panic("未实现") }

func (s *testServer) Config() *config.Config { panic("未实现") }

func (s *testServer) DisableCompress(disable bool) { panic("未实现") }

func (s *testServer) GetRouter(name string) *Router { return s.routers.Router(name) }

func (s *testServer) Location() *time.Location { return time.Local }

func (s *testServer) Logs() *Logs { return s.logs }

func (s *testServer) Name() string { return "test" }

func (s *testServer) NewClient(c *http.Client, sel selector.Selector, mn string, m func(any) ([]byte, error)) *Client {
	return NewClient(c, s.b.codec, sel, mn, m, s.b.requestIDKey, s.unique.String)
}

func (s *testServer) NewContext(w http.ResponseWriter, r *http.Request, route types.Route) *Context {
	return s.b.NewContext(w, r, route)
}

func (s *testServer) NewRouter(name string, matcher RouterMatcher, o ...RouterOption) *Router {
	panic("未实现")
}

func (s *testServer) Now() time.Time { return time.Now().In(s.Location()) }

func (s *testServer) OnClose(f ...func() error) { s.closes = append(s.closes, f...) }

func (s *testServer) ParseTime(layout, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, s.Location())
}

func (s *testServer) RemoveRouter(name string) { panic("未实现") }

func (s *testServer) Routers() []*Router { panic("未实现") }

func (s *testServer) Serve() (err error) { panic("未实现") }

func (s *testServer) State() State { panic("未实现") }

func (s *testServer) UniqueID() string { return s.unique.String() }

func (s *testServer) Uptime() time.Time { panic("未实现") }

func (s *testServer) UseMiddleware(m ...Middleware) { panic("未实现") }

func (s *testServer) Vars() *sync.Map { panic("未实现") }

func (s *testServer) Version() string { return "1.0.0" }

func (s *testServer) CanCompress() bool { return !s.disableCompress }

func (s *testServer) SetCompress(e bool) { s.disableCompress = !e }

func (s *testServer) Problems() *Problems { return s.b.problems }

func (s *testServer) Locale() Locale { return s.locale }

func (s *testServer) Services() *Services { return s.b.Services() }

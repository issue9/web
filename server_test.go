// SPDX-License-Identifier: MIT

package web

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/issue9/assert/v3"
	"github.com/issue9/cache"
	"github.com/issue9/cache/caches/memory"
	"github.com/issue9/config"
	"github.com/issue9/logs/v7"
	"github.com/issue9/mux/v7/group"
	"github.com/issue9/unique/v2"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/internal/locale"
)

var _ Locale = &locale.Locale{}

type testServer struct {
	a               *assert.Assertion
	logs            *logs.Logs
	logBuf          *bytes.Buffer
	unique          *unique.Unique
	cache           cache.Driver
	routers         *group.GroupOf[HandlerFunc]
	b               *InternalContextBuilder
	disableCompress bool
	locale          *locale.Locale
}

type testProblems struct{}

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

	u := unique.NewNumber(100)
	go u.Serve(context.Background())

	cc, _ := memory.New()

	srv := &testServer{
		a:      a,
		logs:   log,
		logBuf: logBuf,
		unique: u,
		cache:  cc,
		locale: locale.New(l, nil, nil),
	}
	srv.b = InternalNewContextBuilder(srv, newCodec(a), header.RequestIDKey)

	return srv
}

func (s *testServer) Cache() cache.Cleanable { return s.cache }

func (s *testServer) Close(shutdownTimeout time.Duration) { panic("未实现") }

func (s *testServer) CompressIsDisable() bool { panic("未实现") }

func (s *testServer) Config() *config.Config { panic("未实现") }

func (s *testServer) DisableCompress(disable bool) { panic("未实现") }

func (s *testServer) GetRouter(name string) *Router { return s.routers.Router(name) }

func (s *testServer) Location() *time.Location { return time.Local }

func (s *testServer) Logs() *Logs { return s.logs }

func (s *testServer) Name() string { return "test" }

func (s *testServer) NewClient(client *http.Client, selector Selector, marshalName string, m func(any) ([]byte, error)) *Client {
	panic("未实现")
}

func (s *testServer) NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return s.b.NewContext(w, r, nil)
}

func (s *testServer) NewRouter(name string, matcher RouterMatcher, o ...RouterOption) *Router {
	panic("未实现")
}

func (s *testServer) Now() time.Time { return time.Now().In(s.Location()) }

func (s *testServer) OnClose(f ...func() error) { panic("未实现") }

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

func (s *testServer) Problems() Problems { return &testProblems{} }

func (s *testServer) Locale() Locale { return s.locale }

func (s *testProblems) Init(pp *RFC7807, id string, p *message.Printer) {
	status, err := strconv.Atoi(id[:3])
	if err != nil {
		panic(err)
	}
	pp.Init(id, id+" title", id+" detail", status)
}

func (s *testProblems) Add(int, ...LocaleProblem) Problems { panic("未实现") }

func (s *testProblems) Visit(func(string, int, LocaleStringer, LocaleStringer)) { panic("未实现") }

func (s *testProblems) Prefix() string { return "" }

func (s *testServer) Services() Services { panic("未实现") }

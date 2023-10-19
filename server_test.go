// SPDX-License-Identifier: MIT

package web

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"context"
	"encoding/json"
	"encoding/xml"
	"io"
	"io/fs"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/issue9/assert/v3"
	"github.com/issue9/config"
	"github.com/issue9/localeutil"
	"github.com/issue9/mux/v7/group"
	"github.com/issue9/mux/v7/types"
	"github.com/issue9/unique/v2"
	"golang.org/x/text/encoding"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web/cache"
	"github.com/issue9/web/cache/caches"
	"github.com/issue9/web/logs"
)

const requestIDKey = "x-request-id"

type testServer struct {
	a          *assert.Assertion
	httpServer *http.Server
	language   language.Tag
	logs       logs.Logs
	logBuf     *bytes.Buffer
	catalog    *catalog.Builder
	unique     *unique.Unique
	cache      cache.Driver
	routers    *group.GroupOf[HandlerFunc]
	printer    *localeutil.Printer
	codec      *testCodec
}

type testCodec struct{}

type testMimetype struct {
	name, problem string
	mb            BuildMarshalFunc
}

func (t *testMimetype) Name(p bool) string {
	if p {
		return t.problem
	}
	return t.name
}

func (t *testMimetype) MarshalBuilder() BuildMarshalFunc { return t.mb }

func (s *testCodec) ContentType(h string) (UnmarshalFunc, encoding.Encoding, error) {
	switch h {
	case "application/json":
		return json.Unmarshal, nil, nil
	case "application/xml":
		return xml.Unmarshal, nil, nil
	default:
		if h != "" {
			return nil, nil, ErrUnsupportedSerialization()
		}
		return json.Unmarshal, nil, nil
	}
}

func (s *testCodec) Accept(h string) Accepter {
	switch h {
	case "application/json", "*/*":
		return &testMimetype{name: h, problem: "application/problem+json", mb: func(*Context) MarshalFunc { return json.Marshal }}
	case "application/xml":
		return &testMimetype{name: h, problem: "application/problem+xml", mb: func(*Context) MarshalFunc { return xml.Marshal }}
	default:
		if h != "" {
			return nil
		}
		return &testMimetype{name: h, problem: "application/problem+json", mb: func(*Context) MarshalFunc { return json.Marshal }}
	}
}

func (s *testCodec) AcceptHeader() string { return "application/json,application/xml" }

func (s *testCodec) ContentEncoding(name string, r io.Reader) (io.ReadCloser, error) {
	switch name {
	case "gzip":
		return gzip.NewReader(r)
	case "deflate":
		return flate.NewReader(r), nil
	default:
		return nil, nil
	}
}

func (s *testCodec) AcceptEncoding(contentType, h string, l logs.Logger) (w CompressorWriterFunc, name string, notAcceptable bool) {
	switch h {
	case "gzip":
		return func(w io.Writer) (io.WriteCloser, error) {
			return gzip.NewWriter(w), nil
		}, "gzip", false
	case "deflate":
		return func(w io.Writer) (io.WriteCloser, error) {
			return flate.NewWriter(w, 3)
		}, "gzip", false
	default:
		return nil, "", h != ""
	}
}

func (s *testCodec) AcceptEncodingHeader() string { return "gzip,deflate" }

func (s *testCodec) SetCompress(enable bool) { panic("未实现") }

func (s *testCodec) CanCompress() bool { panic("未实现") }

func notFound(ctx *Context) Responser { return ctx.NotFound() }

func buildNodeHandle(status int) types.BuildNodeHandleOf[HandlerFunc] {
	return func(n types.Node) HandlerFunc {
		return func(ctx *Context) Responser {
			ctx.Header().Set("Allow", n.AllowHeader())
			if ctx.Request().Method == http.MethodOptions { // OPTIONS 200
				return ResponserFunc(func(ctx *Context) Problem {
					ctx.WriteHeader(http.StatusOK)
					return nil
				})
			}
			return ctx.Problem(strconv.Itoa(status))
		}
	}
}

func (srv *testServer) call(w http.ResponseWriter, r *http.Request, ps types.Route, f HandlerFunc) {
	if ctx := NewContext(srv, w, r, ps, requestIDKey); ctx != nil {
		if resp := f(ctx); resp != nil {
			if p := resp.Apply(ctx); p != nil {
				p.Apply(ctx) // Problem.Apply 始终返回 nil
			}
		}
		ctx.Free()
	}
}

func newTestServer(a *assert.Assertion) *testServer {
	c := catalog.NewBuilder()
	a.NotError(c.SetString(language.SimplifiedChinese, "lang", "cn"))
	a.NotError(c.SetString(language.TraditionalChinese, "lang", "tw"))

	l := language.SimplifiedChinese

	p := message.NewPrinter(l, message.Catalog(c))

	logBuf := new(bytes.Buffer)
	log, err := logs.New(p, &logs.Options{
		Handler:  logs.NewTextHandler(logBuf, os.Stderr),
		Levels:   logs.AllLevels(),
		Location: true,
		Created:  logs.NanoLayout,
	})
	a.NotError(err).NotNil(log)

	u := unique.NewNumber(100)
	go u.Serve(context.Background())

	cc, _ := caches.NewMemory()

	srv := &testServer{
		a:          a,
		httpServer: &http.Server{Addr: ":8080"},
		language:   l,
		logs:       log,
		logBuf:     logBuf,
		catalog:    c,
		unique:     u,
		cache:      cc,
		printer:    p,
	}

	srv.routers = group.NewOf(srv.call, notFound, buildNodeHandle(http.StatusMethodNotAllowed), buildNodeHandle(http.StatusOK))
	srv.httpServer.Handler = srv.routers

	return srv
}

func (s *testServer) AddProblem(id string, status int, title, detail LocaleStringer) {
	panic("未实现")
}

func (s *testServer) Cache() cache.Cleanable { return s.cache }

func (s *testServer) Catalog() *catalog.Builder { return s.catalog }

func (s *testServer) Close(shutdownTimeout time.Duration) {
	s.a.NotError(s.httpServer.Close())
}

func (s *testServer) CompressIsDisable() bool { panic("未实现") }

func (s *testServer) Config() *config.Config { panic("未实现") }

func (s *testServer) DisableCompress(disable bool) { panic("未实现") }

func (s *testServer) GetRouter(name string) *Router { return s.routers.Router(name) }

func (s *testServer) Language() language.Tag { return s.language }

func (s *testServer) LoadLocale(glob string, fsys ...fs.FS) error { panic("未实现") }

func (s *testServer) LocalePrinter() *message.Printer { return s.printer }

func (s *testServer) Location() *time.Location { panic("未实现") }

func (s *testServer) Logs() Logs { return s.logs }

func (s *testServer) Name() string { return "test" }

func (s *testServer) NewClient(client *http.Client, url, marshalName string) *Client {
	panic("未实现")
}

func (s *testServer) NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return NewContext(s, w, r, nil, requestIDKey)
}

func (s *testServer) NewLocalePrinter(tag language.Tag) *message.Printer {
	return message.NewPrinter(tag, message.Catalog(s.Catalog()))
}

func (s *testServer) NewRouter(name string, matcher RouterMatcher, o ...RouterOption) *Router {
	return s.routers.New(name, matcher, o...)
}

func (s *testServer) Now() time.Time { return time.Now().In(s.Location()) }

func (s *testServer) OnClose(f ...func() error) { panic("未实现") }

func (s *testServer) ParseTime(layout, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, s.Location())
}

func (s *testServer) RemoveRouter(name string) { panic("未实现") }

func (s *testServer) Routers() []*Router { panic("未实现") }

func (s *testServer) Serve() (err error) {
	return s.httpServer.ListenAndServe()
}

func (s *testServer) State() State { panic("未实现") }

func (s *testServer) UniqueID() string { return s.unique.String() }

func (s *testServer) Uptime() time.Time { panic("未实现") }

func (s *testServer) UseMiddleware(m ...Middleware) { panic("未实现") }

func (s *testServer) Vars() *sync.Map { panic("未实现") }

func (s *testServer) Version() string { return "1.0.0" }

func (s *testServer) VisitProblems(visit func(prefix, id string, status int, title, detail LocaleStringer)) {
	panic("未实现")
}

func (s *testServer) Codec() Codec { return s.codec }

func (s *testServer) InitProblem(pp *RFC7807, id string, p *message.Printer) {
	status, err := strconv.Atoi(id[:3])
	if err != nil {
		panic(err)
	}
	pp.Init(id, id+" title", id+" detail", status)
}

func (s *testServer) Services() Services { panic("未实现") }

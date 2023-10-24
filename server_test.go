// SPDX-License-Identifier: MIT

package web

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
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
	"github.com/issue9/unique/v2"
	"golang.org/x/text/encoding"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web/cache"
	"github.com/issue9/web/cache/caches"
	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/logs"
)

type testServer struct {
	a        *assert.Assertion
	language language.Tag
	logs     logs.Logs
	logBuf   *bytes.Buffer
	catalog  *catalog.Builder
	unique   *unique.Unique
	cache    cache.Driver
	routers  *group.GroupOf[HandlerFunc]
	printer  *localeutil.Printer
	codec    *testCodec
}

type testCodec struct{}

type testMimetype struct {
	name, problem string
	mb            BuildMarshalFunc
}

type testProblems struct{}

func buildMarshalTest(_ *Context) MarshalFunc {
	return func(v any) ([]byte, error) {
		switch vv := v.(type) {
		case error:
			return nil, vv
		default:
			return nil, ErrUnsupportedSerialization()
		}
	}
}

func unmarshalTest(bs []byte, v any) error {
	return ErrUnsupportedSerialization()
}

func (t *testMimetype) Name(p bool) string {
	if p {
		return t.problem
	}
	return t.name
}

func (t *testMimetype) MarshalBuilder() BuildMarshalFunc { return t.mb }

func (s *testCodec) ContentType(h string) (UnmarshalFunc, encoding.Encoding, error) {
	mime, _ := header.ParseWithParam(h, "charset")
	switch mime {
	case "application/json":
		return json.Unmarshal, nil, nil
	case "application/xml":
		return xml.Unmarshal, nil, nil
	case "application/test":
		return unmarshalTest, nil, nil
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
	case "application/test":
		return &testMimetype{name: h, problem: "application/problem+test", mb: buildMarshalTest}
	case "nil":
		return &testMimetype{name: h, problem: h, mb: nil}
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
		if name != "" {
			return nil, fmt.Errorf("不支持的压缩方法 %s", name)
		}
		return io.NopCloser(r), nil
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
			return flate.NewWriter(w, 8)
		}, "deflate", false
	default:
		return nil, "", h != ""
	}
}

func (s *testCodec) AcceptEncodingHeader() string { return "gzip,deflate" }

func (s *testCodec) SetCompress(enable bool) { panic("未实现") }

func (s *testCodec) CanCompress() bool { panic("未实现") }

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
		a:        a,
		language: l,
		logs:     log,
		logBuf:   logBuf,
		catalog:  c,
		unique:   u,
		cache:    cc,
		printer:  p,
	}

	return srv
}

func (s *testServer) Cache() cache.Cleanable { return s.cache }

func (s *testServer) Catalog() *catalog.Builder { return s.catalog }

func (s *testServer) Close(shutdownTimeout time.Duration) { panic("未实现") }

func (s *testServer) CompressIsDisable() bool { panic("未实现") }

func (s *testServer) Config() *config.Config { panic("未实现") }

func (s *testServer) DisableCompress(disable bool) { panic("未实现") }

func (s *testServer) GetRouter(name string) *Router { return s.routers.Router(name) }

func (s *testServer) Language() language.Tag { return s.language }

func (s *testServer) LoadLocale(glob string, fsys ...fs.FS) error { panic("未实现") }

func (s *testServer) LocalePrinter() *message.Printer { return s.printer }

func (s *testServer) Location() *time.Location { return time.Local }

func (s *testServer) Logs() Logs { return s.logs }

func (s *testServer) Name() string { return "test" }

func (s *testServer) NewClient(client *http.Client, marshalName string, selector Selector) *Client {
	panic("未实现")
}

func (s *testServer) NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return NewContext(s, w, r, nil, header.RequestIDKey)
}

func (s *testServer) NewLocalePrinter(tag language.Tag) *message.Printer {
	return message.NewPrinter(tag, message.Catalog(s.Catalog()))
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

func (s *testServer) Problems() Problems { return &testProblems{} }

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

func (s *testServer) Codec() Codec { return s.codec }

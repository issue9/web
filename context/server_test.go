// SPDX-License-Identifier: MIT

package context

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/issue9/assert"
	"github.com/issue9/cache/memory"
	"github.com/issue9/logs/v2"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/issue9/web/context/mimetype"
	"github.com/issue9/web/context/mimetype/gob"
	"github.com/issue9/web/context/mimetype/mimetypetest"
)

var f201 = func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusCreated)
	_, err := w.Write([]byte("1234567890"))
	if err != nil {
		println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// 声明一个 server 实例
func newServer(a *assert.Assertion) *Server {
	u, err := url.Parse("/root")
	a.NotError(err).NotNil(u)
	srv := NewServer(logs.New(), memory.New(time.Hour), false, false, u)
	a.NotNil(srv)

	// srv.Catalog 默认指向 message.DefaultCatalog
	a.NotError(message.SetString(language.Und, "lang", "und"))
	a.NotError(message.SetString(language.SimplifiedChinese, "lang", "hans"))
	a.NotError(message.SetString(language.TraditionalChinese, "lang", "hant"))

	err = srv.AddMarshals(map[string]mimetype.MarshalFunc{
		"application/json":       json.Marshal,
		"application/xml":        xml.Marshal,
		mimetype.DefaultMimetype: gob.Marshal,
		mimetypetest.Mimetype:    mimetypetest.TextMarshal,
	})
	a.NotError(err)

	err = srv.AddUnmarshals(map[string]mimetype.UnmarshalFunc{
		"application/json":       json.Unmarshal,
		"application/xml":        xml.Unmarshal,
		mimetype.DefaultMimetype: gob.Unmarshal,
		mimetypetest.Mimetype:    mimetypetest.TextUnmarshal,
	})
	a.NotError(err)

	srv.AddMessages(411, map[int]message.Reference{41110: "41110"})

	return srv
}

func TestNewServer(t *testing.T) {
	a := assert.New(t)
	l := logs.New()
	srv := NewServer(l, memory.New(time.Hour), false, false, &url.URL{})
	a.NotNil(srv)
	a.False(srv.Uptime().IsZero())
	a.Equal(l, srv.Logs())
	a.NotNil(srv.Cache())
	a.Equal(srv.Catalog, message.DefaultCatalog)
	a.Equal(srv.Location, time.Local)
	a.Equal(srv.root, "")

	u, err := url.Parse("/root")
	a.NotError(err).NotNil(u)
	srv = NewServer(l, memory.New(time.Hour), false, false, u)
	a.Equal(srv.root, "/root")
}

func TestServer_Vars(t *testing.T) {
	a := assert.New(t)
	srv := newServer(a)

	type (
		t1 int
		t2 int64
		t3 = t2
	)
	var (
		v1 t1 = 1
		v2 t2 = 1
		v3 t3 = 1
	)

	srv.Vars[v1] = 1
	srv.Vars[v2] = 2
	srv.Vars[v3] = 3

	a.Equal(srv.Vars[v1], 1).Equal(srv.Vars[v2], 3)
}

func TestServer_URL_Path(t *testing.T) {
	a := assert.New(t)

	data := []*struct {
		root, input, url, path string
	}{
		{},

		{
			root:  "",
			input: "/abc",
			url:   "/abc",
			path:  "/abc",
		},

		{
			root:  "/",
			input: "/abc/def",
			url:   "/abc/def",
			path:  "/abc/def",
		},

		{
			root:  "https://localhost/",
			input: "/abc/def",
			url:   "https://localhost/abc/def",
			path:  "/abc/def",
		},
		{
			root:  "https://localhost",
			input: "/abc/def",
			url:   "https://localhost/abc/def",
			path:  "/abc/def",
		},
		{
			root:  "https://localhost",
			input: "abc/def",
			url:   "https://localhost/abc/def",
			path:  "/abc/def",
		},

		{
			root:  "https://localhost/",
			input: "",
			url:   "https://localhost",
			path:  "",
		},

		{
			root:  "https://example.com:8080/def/",
			input: "",
			url:   "https://example.com:8080/def",
			path:  "/def",
		},

		{
			root:  "https://example.com:8080/def/",
			input: "abc",
			url:   "https://example.com:8080/def/abc",
			path:  "/def/abc",
		},
	}

	for i, item := range data {
		u, err := url.Parse(item.root)
		a.NotError(err).NotNil(u)
		srv := NewServer(logs.New(), memory.New(time.Hour), false, false, u)
		a.NotNil(srv, "nil at %d", i)

		a.Equal(srv.URL(item.input), item.url, "not equal @%d,v1=%s,v2=%s", i, srv.URL(item.input), item.url)
		a.Equal(srv.Path(item.input), item.path, "not equal @%d,v1=%s,v2=%s", i, srv.Path(item.input), item.path)
	}
}

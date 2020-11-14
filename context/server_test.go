// SPDX-License-Identifier: MIT

package context

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"testing"
	"time"

	"github.com/issue9/assert"
	"github.com/issue9/logs/v2"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/issue9/web/context/contentype"
	"github.com/issue9/web/context/contentype/gob"
	"github.com/issue9/web/context/contentype/mimetypetest"
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
	o := &Options{Root: "/root"}
	srv, err := NewServer(logs.New(), o)
	a.NotError(err).NotNil(srv)

	// srv.Catalog 默认指向 message.DefaultCatalog
	a.NotError(message.SetString(language.Und, "lang", "und"))
	a.NotError(message.SetString(language.SimplifiedChinese, "lang", "hans"))
	a.NotError(message.SetString(language.TraditionalChinese, "lang", "hant"))

	err = srv.Mimetypes().AddMarshals(map[string]contentype.MarshalFunc{
		"application/json":         json.Marshal,
		"application/xml":          xml.Marshal,
		contentype.DefaultMimetype: gob.Marshal,
		mimetypetest.Mimetype:      mimetypetest.TextMarshal,
	})
	a.NotError(err)

	err = srv.Mimetypes().AddUnmarshals(map[string]contentype.UnmarshalFunc{
		"application/json":         json.Unmarshal,
		"application/xml":          xml.Unmarshal,
		contentype.DefaultMimetype: gob.Unmarshal,
		mimetypetest.Mimetype:      mimetypetest.TextUnmarshal,
	})
	a.NotError(err)

	srv.AddMessages(411, map[int]message.Reference{41110: "41110"})

	return srv
}

func TestNewServer(t *testing.T) {
	a := assert.New(t)
	l := logs.New()
	srv, err := NewServer(l, &Options{})
	a.NotError(err).NotNil(srv)
	a.False(srv.Uptime().IsZero())
	a.Equal(l, srv.Logs())
	a.NotNil(srv.Cache())
	a.Equal(srv.catalog, message.DefaultCatalog)
	a.Equal(srv.Location(), time.Local)
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

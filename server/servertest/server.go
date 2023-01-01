// SPDX-License-Identifier: MIT

package servertest

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"os"

	"github.com/issue9/assert/v3"
	"github.com/issue9/localeutil"
	"github.com/issue9/logs/v4"
	"github.com/issue9/term/v3/colors"
	"golang.org/x/text/language"

	"github.com/issue9/web/server"
)

// NewServer 返回功能齐全的 [server.Server] 实例
func NewServer(a *assert.Assertion, o *server.Options) *server.Server {
	s, _ := newServer(a, o)
	return s
}

func newServer(a *assert.Assertion, o *server.Options) (*server.Server, *server.Options) {
	if o == nil {
		o = &server.Options{HTTPServer: &http.Server{Addr: ":8080"}}
	}
	if o.Logs == nil { // 默认重定向到 os.Stderr
		o.Logs = logs.New(logs.NewTermWriter("[15:04:05]", colors.Red, os.Stderr), logs.Caller, logs.Created)
	}

	srv, err := server.New("app", "0.1.0", o)
	a.NotError(err).NotNil(srv)
	a.Equal(srv.Name(), "app").Equal(srv.Version(), "0.1.0")

	// mimetype
	mimetype := srv.Mimetypes()
	mimetype.Add("application/json", server.MarshalJSON, json.Unmarshal, "")
	mimetype.Add("application/xml", server.MarshalXML, xml.Unmarshal, "")
	mimetype.Add("nil", nil, nil, "")
	mimetype.Add("application/test", marshalTest, unmarshalTest, "")
	a.Equal(mimetype.Len(), 4)

	// locale
	b := srv.CatalogBuilder()
	a.NotError(b.SetString(language.Und, "lang", "und"))
	a.NotError(b.SetString(language.SimplifiedChinese, "lang", "hans"))
	a.NotError(b.SetString(language.TraditionalChinese, "lang", "hant"))

	// encoding
	e := srv.Encodings()
	e.Add("gzip", "gzip", server.EncodingGZip(8))
	e.Add("deflate", "deflate", server.EncodingDeflate(8))
	e.Allow("*", "gzip", "deflate")

	srv.Problems().Add("41110", 411, localeutil.Phrase("41110"), localeutil.Phrase("41110"))

	return srv, o
}

func marshalTest(_ *server.Context, v any) ([]byte, error) {
	switch vv := v.(type) {
	case error:
		return nil, vv
	default:
		return nil, server.ErrUnsupported
	}
}

func unmarshalTest(bs []byte, v any) error {
	return server.ErrUnsupported
}

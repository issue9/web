// SPDX-License-Identifier: MIT

package servertest

import (
	"encoding/json"
	"encoding/xml"
	"os"

	"github.com/issue9/assert/v2"
	"github.com/issue9/localeutil"
	"github.com/issue9/logs/v4"
	"github.com/issue9/term/v3/colors"
	"golang.org/x/text/language"

	"github.com/issue9/web/serialization"
	"github.com/issue9/web/serialization/gob"
	"github.com/issue9/web/serialization/text"
	"github.com/issue9/web/server"
)

// NewServer 返回功能齐全的 Server 实例
func NewServer(a *assert.Assertion, o *server.Options) *server.Server {
	s, _ := newServer(a, o)
	return s
}

func newServer(a *assert.Assertion, o *server.Options) (*server.Server, *server.Options) {
	if o == nil {
		o = &server.Options{Port: ":8080"}
	}
	if o.Logs == nil { // 默认重定向到 os.Stderr
		o.Logs = logs.New(logs.NewTermWriter("[15:04:05]", colors.Red, os.Stderr), logs.Caller, logs.Created)
	}

	// encoding
	if o.Encodings == nil {
		o.Encodings = serialization.NewEncodings(o.Logs.ERROR())
		o.Encodings.Add(map[string]serialization.EncodingWriterFunc{
			"gzip":    serialization.GZipWriter,
			"deflate": serialization.DeflateWriter,
		})
	}

	// mimetype
	mimetype := serialization.NewMimetypes(10)
	a.NotError(mimetype.Add(json.Marshal, json.Unmarshal, "application/json"))
	a.NotError(mimetype.Add(xml.Marshal, xml.Unmarshal, "application/xml"))
	a.NotError(mimetype.Add(gob.Marshal, gob.Unmarshal, server.DefaultMimetype))
	a.NotError(mimetype.Add(text.Marshal, text.Unmarshal, text.Mimetype))
	a.NotError(mimetype.Add(nil, nil, "nil"))
	o.Mimetypes = mimetype

	srv, err := server.New("app", "0.1.0", o)
	a.NotError(err).NotNil(srv)
	a.Equal(srv.Name(), "app").Equal(srv.Version(), "0.1.0")

	// locale
	b := srv.Locale().Builder()
	a.NotError(b.SetString(language.Und, "lang", "und"))
	a.NotError(b.SetString(language.SimplifiedChinese, "lang", "hans"))
	a.NotError(b.SetString(language.TraditionalChinese, "lang", "hant"))

	srv.AddResult(411, "41110", localeutil.Phrase("41110"))

	return srv, o
}

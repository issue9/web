// SPDX-License-Identifier: MIT

package servertest

import (
	"net/http"
	"os"

	"github.com/issue9/assert/v3"
	"github.com/issue9/localeutil"
	"github.com/issue9/term/v3/colors"
	"golang.org/x/text/language"

	"github.com/issue9/web/logs"
	"github.com/issue9/web/serializer"
	"github.com/issue9/web/serializer/json"
	"github.com/issue9/web/serializer/xml"
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
		o.Logs = &logs.Options{
			Writer:  logs.NewTermWriter("[15:04:05]", colors.Red, os.Stderr),
			Caller:  true,
			Created: true,
			Levels:  logs.AllLevels(),
		}
	}

	if o.Encodings == nil {
		o.Encodings = []*server.Encoding{
			{Name: "gzip", Builder: server.EncodingGZip(8)},
			{Name: "deflate", Builder: server.EncodingDeflate(8)},
		}
	}
	if o.Mimetypes == nil {
		o.Mimetypes = []*server.Mimetype{
			{Type: "application/json", Marshal: json.Marshal, Unmarshal: json.Unmarshal, ProblemType: json.ProblemMimetype},
			{Type: "application/xml", Marshal: xml.Marshal, Unmarshal: xml.Unmarshal, ProblemType: xml.ProblemMimetype},
			{Type: "application/test", Marshal: marshalTest, Unmarshal: unmarshalTest, ProblemType: ""},
			{Type: "nil", Marshal: nil, Unmarshal: nil, ProblemType: ""},
		}
	}

	srv, err := server.New("app", "0.1.0", o)
	a.NotError(err).NotNil(srv)
	a.Equal(srv.Name(), "app").Equal(srv.Version(), "0.1.0")

	// locale
	b := srv.CatalogBuilder()
	a.NotError(b.SetString(language.Und, "lang", "und"))
	a.NotError(b.SetString(language.SimplifiedChinese, "lang", "hans"))
	a.NotError(b.SetString(language.TraditionalChinese, "lang", "hant"))

	srv.AddProblem("41110", 411, localeutil.Phrase("41110"), localeutil.Phrase("41110"))

	return srv, o
}

func marshalTest(_ *server.Context, v any) ([]byte, error) {
	switch vv := v.(type) {
	case error:
		return nil, vv
	default:
		return nil, serializer.ErrUnsupported()
	}
}

func unmarshalTest(bs []byte, v any) error {
	return serializer.ErrUnsupported()
}

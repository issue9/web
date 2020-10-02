// SPDX-License-Identifier: MIT

package module

import (
	"encoding/json"
	"encoding/xml"
	"log"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/issue9/assert"
	"github.com/issue9/logs/v2"

	"github.com/issue9/web/context"
	"github.com/issue9/web/context/mimetype"
	"github.com/issue9/web/context/mimetype/gob"
)

func newServer(a *assert.Assertion) *Server {
	ctx := context.NewServer(logs.New(), context.DefaultResultBuilder, false, false, &url.URL{})
	a.NotNil(ctx)
	srv, err := NewServer(ctx, "")
	a.NotError(err).NotNil(srv)

	a.NotError(srv.ctxServer.AddMarshals(map[string]mimetype.MarshalFunc{
		"application/json":       json.Marshal,
		"application/xml":        xml.Marshal,
		mimetype.DefaultMimetype: gob.Marshal,
	}))

	a.NotError(srv.ctxServer.AddUnmarshals(map[string]mimetype.UnmarshalFunc{
		"application/json":       json.Unmarshal,
		"application/xml":        xml.Unmarshal,
		mimetype.DefaultMimetype: gob.Unmarshal,
	}))

	a.NotNil(srv.ctxServer).Equal(srv.ctxServer, srv.ctxServer)
	a.NotNil(srv.ctxServer.Logs())

	return srv
}

func TestServer_Init(t *testing.T) {
	a := assert.New(t)
	srv := newServer(a)

	m1 := srv.NewModule("m1", "m1 desc", "m2")
	m1.AddCron("test cron", job, "* * 8 * * *", true)
	m1.AddAt("test cron", job, "2020-01-02 17:55:11", true)

	m2 := srv.NewModule("m2", "m2 desc")
	m2.AddTicker("ticker test", job, 5*time.Second, false, false)

	a.Equal(len(srv.Modules()), 2)

	a.NotError(srv.Init("", log.New(os.Stdout, "[INFO]", 0)))

	// 不能多次调用
	a.Equal(srv.Init("", log.New(os.Stdout, "[INFO]", 0)), ErrInited)
}

// SPDX-License-Identifier: MIT

package module

import (
	"encoding/json"
	"encoding/xml"

	"github.com/issue9/assert"
	"github.com/issue9/logs/v2"

	"github.com/issue9/web/context"
	"github.com/issue9/web/context/mimetype"
	"github.com/issue9/web/context/mimetype/gob"
)

func newServer(a *assert.Assertion) *Server {
	ctx, err := context.NewServer(logs.New(), context.DefaultResultBuilder, false, false, "")
	a.NotError(err).NotNil(ctx)
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

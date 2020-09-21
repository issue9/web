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

func newServer(a *assert.Assertion) *Modules {
	ctx, err := context.NewServer(logs.New(), context.DefaultResultBuilder, false, false, "")
	a.NotError(err).NotNil(ctx)
	ms, err := NewModules(ctx, "")
	a.NotError(err).NotNil(ms)

	a.NotError(ms.ctxServer.AddMarshals(map[string]mimetype.MarshalFunc{
		"application/json":       json.Marshal,
		"application/xml":        xml.Marshal,
		mimetype.DefaultMimetype: gob.Marshal,
	}))

	a.NotError(ms.ctxServer.AddUnmarshals(map[string]mimetype.UnmarshalFunc{
		"application/json":       json.Unmarshal,
		"application/xml":        xml.Unmarshal,
		mimetype.DefaultMimetype: gob.Unmarshal,
	}))

	a.NotNil(ms.ctxServer).Equal(ms.ctxServer, ms.ctxServer)
	a.NotNil(ms.ctxServer.Logs())

	return ms
}

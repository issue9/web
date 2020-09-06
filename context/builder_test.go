// SPDX-License-Identifier: MIT

package context

import (
	"encoding/json"
	"encoding/xml"

	"github.com/issue9/assert"
	"github.com/issue9/logs/v2"
	"github.com/issue9/middleware/recovery/errorhandler"

	"github.com/issue9/web/context/mimetype"
	"github.com/issue9/web/context/mimetype/gob"
	"github.com/issue9/web/context/mimetype/mimetypetest"
)

// 声明一个 builder 实例
func newBuilder(a *assert.Assertion) *Builder {
	b := NewBuilder(DefaultResultBuilder)
	a.NotNil(b)

	err := b.AddMarshals(map[string]mimetype.MarshalFunc{
		"application/json":       json.Marshal,
		"application/xml":        xml.Marshal,
		mimetype.DefaultMimetype: gob.Marshal,
		mimetypetest.Mimetype:    mimetypetest.TextMarshal,
	})
	a.NotError(err)

	err = b.AddUnmarshals(map[string]mimetype.UnmarshalFunc{
		"application/json":       json.Unmarshal,
		"application/xml":        xml.Unmarshal,
		mimetype.DefaultMimetype: gob.Unmarshal,
		mimetypetest.Mimetype:    mimetypetest.TextUnmarshal,
	})
	a.NotError(err)

	b.ErrorHandlers = errorhandler.New()
	b.Logs = logs.New()
	b.ContextInterceptor = func(ctx *Context) {}
	return b
}

// SPDX-License-Identifier: MIT

package context

import (
	"encoding/json"
	"encoding/xml"

	"github.com/issue9/assert"
	"github.com/issue9/logs/v2"
	"github.com/issue9/middleware/recovery/errorhandler"

	"github.com/issue9/web/mimetype"
	"github.com/issue9/web/mimetype/gob"
	"github.com/issue9/web/mimetype/mimetypetest"
	"github.com/issue9/web/result"
)

type builder struct {
	errorHandlers *errorhandler.ErrorHandler
	mimetypes     *mimetype.Mimetypes
	logs          *logs.Logs
	results       *result.Results
}

func (b *builder) ErrorHandlers() *errorhandler.ErrorHandler {
	return b.errorHandlers
}

func (b *builder) Mimetypes() *mimetype.Mimetypes {
	return b.mimetypes
}

func (b *builder) Logs() *logs.Logs {
	return b.logs
}

func (b *builder) Results() *result.Results {
	return b.results
}

func (b *builder) ContextInterceptor(ctx *Context) {
}

// 声明一个 builder 实例
func newBuilder(a *assert.Assertion) Builder {
	mt := mimetype.New()
	err := mt.AddMarshals(map[string]mimetype.MarshalFunc{
		"application/json":       json.Marshal,
		"application/xml":        xml.Marshal,
		mimetype.DefaultMimetype: gob.Marshal,
		mimetypetest.Mimetype:    mimetypetest.TextMarshal,
	})
	a.NotError(err)

	err = mt.AddUnmarshals(map[string]mimetype.UnmarshalFunc{
		"application/json":       json.Unmarshal,
		"application/xml":        xml.Unmarshal,
		mimetype.DefaultMimetype: gob.Unmarshal,
		mimetypetest.Mimetype:    mimetypetest.TextUnmarshal,
	})
	a.NotError(err)

	return &builder{
		errorHandlers: errorhandler.New(),
		mimetypes:     mt,
		logs:          logs.New(),
		results:       result.NewResults(result.DefaultResultBuilder),
	}
}

// SPDX-License-Identifier: MIT

package parser

import (
	"context"
	"testing"

	"github.com/issue9/assert/v3"

	"github.com/issue9/web/cmd/web/internal/restdoc/logger"
	"github.com/issue9/web/cmd/web/internal/restdoc/logger/loggertest"
)

func TestParser(t *testing.T) {
	a := assert.New(t, false)
	l := loggertest.New()
	p := New(l.Logger)

	p.AddDir(context.Background(), "./testdata", true)
	d := p.OpenAPI(context.Background())
	a.NotNil(d).
		Length(l.Entries[logger.GoSyntax], 0).
		Length(l.Entries[logger.Cancelled], 0).
		Length(l.Entries[logger.DocSyntax], 0).
		Length(l.Entries[logger.Unknown], 0)

	a.NotNil(d.Info).Equal(d.Info.Version, "1.0")

	login := d.Paths["/login"].Post
	a.NotNil(login).
		Length(login.Parameters, 3).
		NotNil(login.RequestBody).
		Length(login.Responses, 5) // 包含默认的 default
}

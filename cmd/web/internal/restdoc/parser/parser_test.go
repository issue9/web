// SPDX-License-Identifier: MIT

package parser

import (
	"context"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/web/logs"

	"github.com/issue9/web/cmd/web/internal/restdoc/logger/loggertest"
)

func TestParser(t *testing.T) {
	a := assert.New(t, false)
	l := loggertest.New(a)
	p := New(l.Logger)

	p.AddDir(context.Background(), "./testdata", true)
	d := p.OpenAPI(context.Background())
	a.NotNil(d).
		Length(l.Records[logs.Error], 0).
		Length(l.Records[logs.Warn], 0).
		Length(l.Records[logs.Info], 2)

	a.NotNil(d.Info).Equal(d.Info.Version, "1.0")

	login := d.Paths["/login"].Post
	a.NotNil(login).
		Length(login.Parameters, 3).
		NotNil(login.RequestBody).
		Length(login.Responses, 5) // 包含默认的 default
}

// SPDX-License-Identifier: MIT

package parser

import (
	"context"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
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

	a.NotNil(d.Info).Equal(d.Info.Version, "1.0.0")

	login := d.Paths["/login"].Post
	a.NotNil(login).
		Length(login.Parameters, 4).
		Equal(login.Parameters[3].Value.Name, "type").
		Equal(login.Parameters[3].Value.Schema.Value.Type, openapi3.TypeString).
		Equal(login.Parameters[2].Value.Schema.Value.Type, openapi3.TypeArray).
		Equal(login.Parameters[2].Value.Schema.Value.Items.Value.Type, openapi3.TypeInteger).
		NotNil(login.RequestBody).
		Length(login.Responses, 5). // 包含默认的 default
		Length(login.Callbacks, 1).
		NotNil((*login.Callbacks["onData"].Value)["{$request.query.url}"].Post)

	a.NotError(p.SaveAs(context.Background(), "./testdata/openapi.out.yaml"))
}

func TestIsIgnoreTag(t *testing.T) {
	a := assert.New(t, false)

	a.False(isIgnoreTag(nil, "t1"))
	a.True(isIgnoreTag([]string{"t10"}, "t1"))
	a.False(isIgnoreTag([]string{"t10"}, "t10"))
	a.False(isIgnoreTag([]string{"t10"}, "t10", "t1"))
	a.False(isIgnoreTag([]string{"t1", "t10"}, "t10", "t1"))
}

// SPDX-License-Identifier: MIT

package parser

import (
	"context"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/issue9/assert/v3"
	"github.com/issue9/logs/v7"

	"github.com/issue9/web/cmd/web/restdoc/logger/loggertest"
)

func TestParser_Parse(t *testing.T) {
	a := assert.New(t, false)
	l := loggertest.New(a)
	p := New(l.Logger, "/prefix", nil)

	p.AddDir(context.Background(), "./testdata", true)
	d := p.Parse(context.Background())
	a.NotNil(d).
		Length(l.Records[logs.LevelError], 0).
		Length(l.Records[logs.LevelWarn], 0).
		Length(l.Records[logs.LevelInfo], 0)

	a.NotNil(d.Doc().Info).
		Equal(d.Doc().Info.Version, "1.0.0")

	login := d.Doc().Paths.Find("/prefix/login").Post
	a.NotNil(login).
		Length(login.Parameters, 6).
		Equal(login.Parameters[3].Value.Name, "sex").
		NotEmpty(login.Parameters[3].Value.Schema.Ref).
		Equal(login.Parameters[4].Value.Name, "type").
		Equal(login.Parameters[4].Value.Schema.Value.Type, openapi3.TypeString).
		Equal(login.Parameters[2].Value.Schema.Value.Type, openapi3.TypeArray).
		Equal(login.Parameters[2].Value.Schema.Value.Items.Value.Type, openapi3.TypeInteger).
		NotNil(login.RequestBody).
		Equal(login.Responses.Len(), 5). // 包含默认的 default
		Length(login.Callbacks, 1).
		NotNil((login.Callbacks["onData"].Value).Value("{$request.query.url}").Post)

	a.NotError(d.SaveAs("./testdata/openapi.out.yaml"))
	a.NotError(d.SaveAs("./testdata/openapi.out.json"))
}

func TestIsIgnoreTag(t *testing.T) {
	a := assert.New(t, false)

	p := &Parser{}
	a.False(p.isIgnoreTag("t1"))

	p = &Parser{tags: []string{"t10"}}
	a.True(p.isIgnoreTag("t1"))

	p = &Parser{tags: []string{"t10"}}
	a.False(p.isIgnoreTag("t10"))

	p = &Parser{tags: []string{"t10"}}
	a.False(p.isIgnoreTag("t10", "t1"))

	p = &Parser{tags: []string{"t10", "t1"}}
	a.False(p.isIgnoreTag("t10", "t1"))
}

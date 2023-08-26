// SPDX-License-Identifier: MIT

package parser

import (
	"context"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/issue9/assert/v3"
	"github.com/issue9/web/logs"

	"github.com/issue9/web/cmd/web/restdoc/logger/loggertest"
)

func TestParser(t *testing.T) {
	a := assert.New(t, false)
	l := loggertest.New(a)
	p := New(l.Logger, "/prefix", nil)

	p.AddDir(context.Background(), "./testdata", true)
	d := p.Parse(context.Background())
	a.NotNil(d).
		Length(l.Records[logs.Error], 0).
		Length(l.Records[logs.Warn], 0).
		Length(l.Records[logs.Info], 0)

	a.NotNil(d.Doc().Info).
		Equal(d.Doc().Info.Version, "1.0.0")

	login := d.Doc().Paths["/prefix/login"].Post
	a.NotNil(login).
		Length(login.Parameters, 6).
		Equal(login.Parameters[3].Value.Name, "sex").
		NotEmpty(login.Parameters[3].Value.Schema.Ref).
		Equal(login.Parameters[4].Value.Name, "type").
		Equal(login.Parameters[4].Value.Schema.Value.Type, openapi3.TypeString).
		Equal(login.Parameters[2].Value.Schema.Value.Type, openapi3.TypeArray).
		Equal(login.Parameters[2].Value.Schema.Value.Items.Value.Type, openapi3.TypeInteger).
		NotNil(login.RequestBody).
		Length(login.Responses, 4). // 包含默认的 default
		Length(login.Callbacks, 1).
		NotNil((*login.Callbacks["onData"].Value)["{$request.query.url}"].Post)

	doc := p.Parse(context.Background())
	a.NotNil(doc)
	a.NotError(doc.SaveAs("./testdata/openapi.out.yaml"))
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

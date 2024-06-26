// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package parser

import (
	"context"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/issue9/assert/v4"
	"github.com/issue9/logs/v7"

	"github.com/issue9/web/cmd/web/restdoc/logger/loggertest"
	"github.com/issue9/web/cmd/web/restdoc/openapi"
)

func TestBuildPath(t *testing.T) {
	a := assert.New(t, false)

	path := "github.com/issue9/web"
	a.Equal(buildPath(path, "github.com/issue9/web.Type"), "github.com/issue9/web.Type").
		Equal(buildPath(path, "abc"), path+".abc").
		Equal(buildPath(path, openapi.ComponentSchemaPrefix+"/abc"), openapi.ComponentSchemaPrefix+"/abc").
		Equal(buildPath(path, "[]abc"), "[]"+path+".abc").
		Equal(buildPath(path, "[]*abc"), "[]*"+path+".abc").
		Equal(buildPath(path, "[]"+path+".abc"), "[]"+path+".abc").
		Equal(buildPath(path, "[5]*abc"), "[5]*"+path+".abc").
		Equal(buildPath(path, "[]abc[int]"), "[]"+path+".abc["+path+".int]").
		Equal(buildPath(path, "[]abc[int,S]"), "[]"+path+".abc["+path+".int,"+path+".S]")

	a.Equal(buildPath(path, "[5x]*abc"), path+".[5x]*abc").
		Equal(buildPath(path, "[*]abc"), path+".[*]abc").
		Equal(buildPath(path, "[[]abc"), path+".[[]abc").
		Equal(buildPath(path, "[]]abc"), path+".[]]abc").
		Equal(buildPath(path, "5abc"), path+".5abc")
}

func TestParser_Parse(t *testing.T) {
	a := assert.New(t, false)
	l := loggertest.New(a)
	p := New(l.Logger, "/prefix", nil)

	p.AddDir(context.Background(), "./testdata", true)
	d := p.Parse(context.Background())
	a.NotNil(d).
		Length(l.Records[logs.LevelError], 0).
		Length(l.Records[logs.LevelWarn], 0).
		Length(l.Records[logs.LevelInfo], 2) // scan dir / add api 的提示

	a.NotNil(d.Doc().Info).
		Equal(d.Doc().Info.Version, "1.0.0")

	login := d.Doc().Paths.Find("/prefix/login").Post
	a.NotNil(login).
		Length(login.Parameters, 6).
		Equal(login.Parameters[3].Value.Name, "sex").
		Empty(login.Parameters[3].Value.Schema.Ref). // 查询参数不保存在 schema，也就没有 Ref 的必要
		Equal(login.Parameters[4].Value.Name, "type").
		True(login.Parameters[4].Value.Schema.Value.Type.Is(openapi3.TypeString)).
		True(login.Parameters[2].Value.Schema.Value.Type.Is(openapi3.TypeArray)).
		True(login.Parameters[2].Value.Schema.Value.Items.Value.Type.Is(openapi3.TypeInteger)).
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

// SPDX-License-Identifier: MIT

package parser

import (
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/web/logs"

	"github.com/issue9/web/cmd/web/internal/restdoc/logger/loggertest"
	"github.com/issue9/web/cmd/web/internal/restdoc/schema"
)

func TestRESTDoc_parseRESTDoc(t *testing.T) {
	a := assert.New(t, false)

	l := loggertest.New(a)
	p := New(l.Logger)
	d := schema.NewOpenAPI("3")
	lines := []string{
		"@version 1.0.0",
		"@tag user user tag desc ",
		"@server https://example.com/v1 v1 desc",
		"@tag admin admin tag desc",
		"@server https://example.com/v2 v2 desc",
		"@license mit https://example.com/license",
		"@contact name https://example.com x@example.com",
		"@term https://example.com/term",
		"@media application/json application/xml",
		"@doc https://doc.example.com",
		"@scy-http http-security bearer format http bearer auth",
		"@scy-apikey apikey-security key header apikey header auth",
		"@scy-openid openid-security https://example.com/openid openid auth",
		"",
		"# markdown desc",
		"line 2",
	}
	p.parseRESTDoc(d, "restdoc example", "github.com/issue9/web", lines, 5, "example.go")

	a.Equal(0, l.Count()).
		Length(d.Tags, 2).Equal(d.Tags[0].Description, "user tag desc").
		Length(d.Servers, 2).
		Equal(d.Info.License.Name, "mit").
		Equal(d.Info.TermsOfService, "https://example.com/term").
		Equal(d.Info.Contact.Name, "name").
		Equal(d.Info.Description, "# markdown desc\nline 2").
		Equal(p.media, []string{"application/json", "application/xml"}).
		Equal(d.ExternalDocs.URL, "https://doc.example.com")

	http := d.Components.SecuritySchemes["http-security"]
	a.NotNil(http).
		Equal(http.Value.Scheme, "bearer").
		Equal(http.Value.BearerFormat, "format").
		Equal(http.Value.Description, "http bearer auth")

	apikey := d.Components.SecuritySchemes["apikey-security"]
	a.NotNil(apikey).
		Equal(apikey.Value.Name, "key").
		Equal(apikey.Value.In, "header").
		Equal(apikey.Value.Description, "apikey header auth")

	openid := d.Components.SecuritySchemes["openid-security"]
	a.NotNil(openid).
		Equal(openid.Value.OpenIdConnectUrl, "https://example.com/openid").
		Equal(openid.Value.Description, "openid auth")

	// 测试行号是否正确
	l = loggertest.New(a)
	p = New(l.Logger)
	d = schema.NewOpenAPI("3")
	lines = []string{
		"@version 1.0.0",
		"@tag user user tag desc",
		"@server",
		"",
		"# markdown desc",
		"line 2",
	}
	p.parseRESTDoc(d, "restdoc example", "github.com/issue9/web", lines, 5, "example.go")

	a.Equal(1, l.Count()).
		Length(d.Tags, 1).
		Equal(d.Info.Description, "# markdown desc\nline 2").
		Contains(l.Records[logs.Error][0], "example.go:8")
}

func TestBuildContact(t *testing.T) {
	a := assert.New(t, false)

	c := buildContact([]string{"name"})
	a.Equal(c.Name, "name").
		Empty(c.Email).
		Empty(c.URL)

	c = buildContact([]string{"https://example.com"})
	a.Equal(c.URL, "https://example.com").
		Empty(c.Email).
		Empty(c.Name)

	c = buildContact([]string{"x@example.com"})
	a.Equal(c.Email, "x@example.com").
		Empty(c.URL).
		Empty(c.Name)

	c = buildContact([]string{"x@example.com", "name"})
	a.Equal(c.Email, "x@example.com").
		Empty(c.URL).
		Equal(c.Name, "name")

	c = buildContact([]string{"x@example.com", "name", "https://example.com"})
	a.Equal(c.Email, "x@example.com").
		Equal(c.URL, "https://example.com").
		Equal(c.Name, "name")
}

func TestParseOpenAPI(t *testing.T) {
	a := assert.New(t, false)

	l := loggertest.New(a)
	p := New(l.Logger)
	d := schema.NewOpenAPI("3.1.0")

	p.parseOpenAPI(d, "./testdata/openapi.yaml", "test.go", 5)
	a.Nil(d.Info).
		Length(d.Paths, 1)
}

// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"testing"

	"github.com/issue9/assert/v4"

	"github.com/issue9/web"
	"github.com/issue9/web/mimetype/cbor"
	"github.com/issue9/web/mimetype/json"
)

func TestWithOptions(t *testing.T) {
	a := assert.New(t, false)

	ss := newServer(a)
	d := New(ss, web.Phrase("title"), WithOptions(WithHeadMethod(true), WithOptionsMethod(true)))
	a.True(d.enableHead).
		True(d.enableOptions)
}

func TestWithHTML(t *testing.T) {
	a := assert.New(t, false)

	ss := newServer(a)
	d := New(ss, web.Phrase("desc"), WithHTML("tpl", "./dist", "./dist/favicon.png"))
	a.Equal(d.assetsURL, "./dist").
		Equal(d.templateName, "tpl").
		Equal(d.logo, "./dist/favicon.png")
}

func TestNewLicense(t *testing.T) {
	a := assert.New(t, false)
	ss := newServer(a)

	d := New(ss, web.Phrase("abc"), WithLicense("name", "id"))
	a.Equal(d.info.license, &licenseRenderer{Name: "name", Identifier: "id"})

	d = New(ss, web.Phrase("abc"), WithLicense("name", "https://example.com"))
	a.Equal(d.info.license, &licenseRenderer{Name: "name", URL: "https://example.com"})

	d = New(ss, web.Phrase("abc"), WithLicense("name", ""))
	a.Equal(d.info.license, &licenseRenderer{Name: "name"})

	a.PanicString(func() {
		WithLicense("", "id")
	}, "参数 name 不能为空")
}

func TestWithResponse(t *testing.T) {
	a := assert.New(t, false)
	ss := newServer(a)

	a.PanicString(func() {
		New(ss, web.Phrase("desc"), WithResponse(&Response{}, "400"))
	}, "必须存在 ref")

	d := New(ss, web.Phrase("desc"),
		WithResponse(&Response{Ref: &Ref{Ref: "400"}}, "400"),
		WithResponse(&Response{Ref: &Ref{Ref: "500"}}, "500"),
	)
	a.NotNil(d).
		Length(d.components.responses, 2).
		Equal(d.responses["400"], "400").
		Equal(d.responses["500"], "500")
}

func TestWithProblemResponse(t *testing.T) {
	a := assert.New(t, false)
	ss := newServer(a)

	d := New(ss, web.Phrase("desc"), WithProblemResponse())
	a.NotNil(d).
		Length(d.responses, 2).
		Length(d.components.responses, 1).
		Equal(d.responses["4XX"], "problem").
		Equal(d.responses["5XX"], "problem")
}

func TestWithMediaType(t *testing.T) {
	a := assert.New(t, false)
	ss := newServer(a)

	d := New(ss, web.Phrase("desc"),
		WithMediaType(json.Mimetype),
	)
	a.NotNil(d).
		Length(d.mediaTypes, 1)

	a.PanicString(func() {
		New(ss, web.Phrase("desc"),
			WithMediaType(json.Mimetype, cbor.Mimetype),
		)
	}, "不支持 application/cbor 媒体类型")
}

func TestWithCallback(t *testing.T) {
	a := assert.New(t, false)
	ss := newServer(a)

	a.PanicString(func() {
		New(ss, web.Phrase("desc"), WithCallback(&Callback{}))
	}, "必须存在 ref")

	a.PanicString(func() {
		New(ss, web.Phrase("desc"), WithCallback(&Callback{Ref: &Ref{Ref: "1"}}))
	}, "Callback 不能为空")

	d := New(ss, web.Phrase("desc"),
		WithCallback(&Callback{Ref: &Ref{Ref: "1"}, Callback: map[string]*PathItem{"path": {}}}),
		WithCallback(&Callback{Ref: &Ref{Ref: "2"}, Callback: map[string]*PathItem{"path": {}}}),
	)
	a.NotNil(d).
		Length(d.components.callbacks, 2)
}

func TestWithHeader(t *testing.T) {
	a := assert.New(t, false)
	ss := newServer(a)

	a.PanicString(func() {
		New(ss, web.Phrase("desc"), WithHeader(true, &Parameter{Name: "h1", Schema: &Schema{Type: TypeString}}))
	}, "必须存在 ref")

	d := New(ss, web.Phrase("desc"),
		WithHeader(true, &Parameter{Ref: &Ref{Ref: "1"}, Name: "h1", Schema: &Schema{Type: TypeString}}),
		WithHeader(true, &Parameter{Ref: &Ref{Ref: "2"}, Name: "h2", Schema: &Schema{Type: TypeString}}),
	)
	a.NotNil(d).
		Length(d.components.headers, 2).
		Length(d.headers, 2)
}

func TestWithCookie(t *testing.T) {
	a := assert.New(t, false)
	ss := newServer(a)

	a.PanicString(func() {
		New(ss, web.Phrase("desc"), WithCookie(true, &Parameter{Name: "c1", Schema: &Schema{Type: TypeString}}))
	}, "必须存在 ref")

	d := New(ss, web.Phrase("desc"),
		WithCookie(true, &Parameter{Ref: &Ref{Ref: "1"}}),
		WithCookie(true, &Parameter{Ref: &Ref{Ref: "2"}}),
	)
	a.NotNil(d).
		Length(d.components.cookies, 2).
		Length(d.cookies, 2)
}

func TestWithDescription(t *testing.T) {
	a := assert.New(t, false)
	ss := newServer(a)
	d := New(ss, web.Phrase("title"), WithDescription(web.Phrase("lang"), web.Phrase("desc")))

	a.Equal(d.info.summary, "lang").
		Equal(d.info.description, "desc")
}

func TestWithServer(t *testing.T) {
	a := assert.New(t, false)
	ss := newServer(a)
	d := New(ss, web.Phrase("title"),
		WithServer("https://example.com/s1", web.Phrase("s1")),
		WithServer("https://example.com/s2/{v1}", web.Phrase("s2"), &ServerVariable{Name: "v1", Default: "1"}),
	)

	a.Length(d.servers, 2).
		Empty(d.servers[0].Variables).
		Length(d.servers[1].Variables, 1)
}

func TestWithTag(t *testing.T) {
	a := assert.New(t, false)
	ss := newServer(a)
	d := New(ss, web.Phrase("title"),
		WithTag("t1", web.Phrase("t1"), "", nil),
		WithTag("t2", web.Phrase("t2"), "https://example.com", web.Phrase("desc")),
	)

	a.Length(d.tags, 2).
		Equal(d.tags[0].name, "t1").
		Equal(d.tags[1].externalDocs.URL, "https://example.com")
}

func TestWithSecurityScheme(t *testing.T) {
	a := assert.New(t, false)
	ss := newServer(a)
	d := New(ss, web.Phrase("title"),
		WithSecurityScheme(&SecurityScheme{
			ID:     "http1",
			Type:   "http",
			Scheme: "basic",
		}, []string{}, []string{"abc"}),
	)

	a.Length(d.components.securitySchemes, 1).
		NotNil(d.components.securitySchemes["http1"]).
		Length(d.security, 2).
		Equal(d.security[0], &SecurityRequirement{Name: "http1", Scopes: []string{}}).
		Equal(d.security[1], &SecurityRequirement{Name: "http1", Scopes: []string{"abc"}})
}

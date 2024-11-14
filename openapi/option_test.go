// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"testing"

	"github.com/issue9/assert/v4"

	"github.com/issue9/web"
	"github.com/issue9/web/mimetype/cbor"
	"github.com/issue9/web/mimetype/html"
	"github.com/issue9/web/mimetype/json"
)

func TestWithHTML(t *testing.T) {
	a := assert.New(t, false)

	d := New("0.1", web.Phrase("desc"), WithHTML("tpl", "./openapi.yaml"))
	a.Equal(d.dataURL, "./openapi.yaml").
		Equal(d.templateName, "tpl")
}

func TestWithResponse(t *testing.T) {
	a := assert.New(t, false)

	a.PanicString(func() {
		New("0.1", web.Phrase("desc"), WithResponse(400, &Response{}))
	}, "必须存在 ref")

	d := New("0.1", web.Phrase("desc"),
		WithResponse(400, &Response{Ref: &Ref{Ref: "400"}}),
		WithResponse(500, &Response{Ref: &Ref{Ref: "500"}}),
	)
	a.NotNil(d).
		Length(d.components.responses, 2).
		Equal(d.responses[400], "400").
		Equal(d.responses[500], "500")
}

func TestWithMediaType(t *testing.T) {
	a := assert.New(t, false)

	d := New("0.1", web.Phrase("desc"),
		WithMediaType(json.Mimetype, cbor.Mimetype),
		WithMediaType(html.Mimetype, cbor.Mimetype),
	)
	a.NotNil(d).
		Length(d.mediaTypes, 3)
}

func TestWithHeader(t *testing.T) {
	a := assert.New(t, false)

	a.PanicString(func() {
		New("0.1", web.Phrase("desc"), WithHeader(&Parameter{}))
	}, "必须存在 ref")

	d := New("0.1", web.Phrase("desc"),
		WithHeader(&Parameter{Ref: &Ref{Ref: "1"}}),
		WithHeader(&Parameter{Ref: &Ref{Ref: "2"}}),
	)
	a.NotNil(d).
		Length(d.components.headers, 2).
		Length(d.headers, 2)
}

func TestWithCookie(t *testing.T) {
	a := assert.New(t, false)

	a.PanicString(func() {
		New("0.1", web.Phrase("desc"), WithCookie(&Parameter{}))
	}, "必须存在 ref")

	d := New("0.1", web.Phrase("desc"),
		WithCookie(&Parameter{Ref: &Ref{Ref: "1"}}),
		WithCookie(&Parameter{Ref: &Ref{Ref: "2"}}),
	)
	a.NotNil(d).
		Length(d.components.cookies, 2).
		Length(d.cookies, 2)
}

func TestWithDescription(t *testing.T) {
	a := assert.New(t, false)
	d := New("0.1", web.Phrase("title"), WithDescription(web.Phrase("lang"), web.Phrase("desc")))

	a.Equal(d.info.summary, "lang").
		Equal(d.info.description, "desc")
}

func TestWithServer(t *testing.T) {
	a := assert.New(t, false)
	d := New("0.1", web.Phrase("title"),
		WithServer("https://example.com/s1", web.Phrase("s1")),
		WithServer("https://example.com/s2/{v1}", web.Phrase("s2"), &ServerVariable{Name: "v1", Default: "1"}),
	)

	a.Length(d.servers, 2).
		Empty(d.servers[0].Variables).
		Length(d.servers[1].Variables, 1)
}

func TestWithTag(t *testing.T) {
	a := assert.New(t, false)
	d := New("0.1", web.Phrase("title"),
		WithTag("t1", web.Phrase("t1"), "", nil),
		WithTag("t2", web.Phrase("t2"), "https://example.com", web.Phrase("desc")),
	)

	a.Length(d.tags, 2).
		Equal(d.tags[0].name, "t1").
		Equal(d.tags[1].externalDocs.URL, "https://example.com")
}

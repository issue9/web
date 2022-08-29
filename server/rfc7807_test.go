// SPDX-License-Identifier: MIT

package server

import (
	"encoding/json"
	"encoding/xml"
	"testing"

	"github.com/issue9/assert/v3"

	"github.com/issue9/web/serializer/form"
)

var (
	_ BuildProblemFunc = RFC7807Builder

	_ Problem        = &rfc7807{}
	_ json.Marshaler = &rfc7807{}
	_ xml.Marshaler  = &rfc7807{}
	_ form.Marshaler = &rfc7807{}
)

func TestRFC7807Builder(t *testing.T) {
	a := assert.New(t, false)
	p := RFC7807Builder("id", "title", 400)
	a.NotNil(p)
	p.With("instance", "https://example.com/instance/1")

	a.PanicString(func() {
		p.With("instance", "instance")
	}, "存在同名的参数")
	a.PanicString(func() {
		p.With("type", "1111")
	}, "存在同名的参数")

	pp, ok := p.(*rfc7807)
	a.True(ok).NotNil(pp)
	a.Equal(pp.vals[3], "https://example.com/instance/1").
		Equal(pp.keys[3], "instance").
		Equal(pp.status, 400).
		Equal(pp.vals[0], "id").
		Equal(pp.vals[1], "title").
		Equal(pp.vals[2], 400)
}

func TestRFC7807_Marshal(t *testing.T) {
	a := assert.New(t, false)

	// NOTE: title 因为未调用 apply，所以未翻译，为空。
	p1 := RFC7807Builder("400", "bad request", 200)
	p2 := RFC7807Builder("400", "bad request", 400)
	p2.AddParam("n1", "r1")
	p2.With("detail", "detail")
	p2.With("array", []string{"a", "bc"})
	p2.With("object", &struct{ X string }{X: "x"})

	t.Run("JSON", func(t *testing.T) {
		data, err := json.Marshal(p1)
		a.NotError(err).
			Equal(string(data), `{"type":"400","title":"bad request","status":200}`)

		data, err = json.Marshal(p2)
		a.NotError(err).
			Equal(string(data), `{"type":"400","title":"bad request","status":400,"detail":"detail","array":["a","bc"],"object":{"X":"x"},"params":[{"name":"n1","reason":"r1"}]}`)
	})

	t.Run("XML", func(t *testing.T) {
		data, err := xml.Marshal(p1)
		a.NotError(err).
			Equal(string(data), `<problem xmlns="urn:ietf:rfc:7807"><type>400</type><title>bad request</title><status>200</status></problem>`)

		data, err = xml.Marshal(p2)
		a.NotError(err).
			Equal(string(data), `<problem xmlns="urn:ietf:rfc:7807"><type>400</type><title>bad request</title><status>400</status><detail>detail</detail><array><i>a</i><i>bc</i></array><object><X>x</X></object><params><i><name>n1</name><reason>r1</reason></i></params></problem>`)
	})

	t.Run("Form", func(t *testing.T) {
		data, err := form.Marshal(p1)
		a.NotError(err).
			Equal(string(data), `status=200&title=bad+request&type=400`)

		data, err = form.Marshal(p2)
		a.NotError(err).
			Equal(string(data), `detail=detail&params%5B0%5D.name=n1&params%5B0%5D.reason=r1&status=400&title=bad+request&type=400`)
	})
}

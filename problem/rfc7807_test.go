// SPDX-License-Identifier: MIT

package problem

import (
	"encoding/json"
	"encoding/xml"
	"testing"

	"github.com/issue9/assert/v2"
)

var (
	_ BuildFunc = RFC7807Builder

	_ Problem        = &rfc7807{}
	_ json.Marshaler = &rfc7807{}
	_ xml.Marshaler  = &rfc7807{}
)

func TestRFC7807Builder(t *testing.T) {
	a := assert.New(t, false)
	p := RFC7807Builder("id", "title", 400)
	a.NotNil(p)
	p.With("instance", "https://example.com/instance/1")

	pp, ok := p.(*rfc7807)
	a.True(ok).NotNil(pp)
	a.Equal(pp.vals[0], "https://example.com/instance/1").
		Equal(pp.keys[0], "instance").
		Equal(pp.status, 400).
		Equal(pp.typ, "id")
}

func TestRFC7807_Marshal(t *testing.T) {
	a := assert.New(t, false)

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
			Equal(string(data), `{"type":"400","title":"bad request","status":400,"params":[{"name":"n1","reason":"r1"}],"detail":"detail","array":["a","bc"],"object":{"X":"x"}}`)
	})

	t.Run("XML", func(t *testing.T) {
		data, err := xml.Marshal(p1)
		a.NotError(err).
			Equal(string(data), `<problem xmlns="urn:ietf:rfc:7807"><type>400</type><title>bad request</title><status>200</status></problem>`)

		data, err = xml.Marshal(p2)
		a.NotError(err).
			Equal(string(data), `<problem xmlns="urn:ietf:rfc:7807"><type>400</type><title>bad request</title><status>400</status><params><i><name>n1</name><reason>r1</reason></i></params><detail>detail</detail><array><i>a</i><i>bc</i></array><object><X>x</X></object></problem>`)
	})
}

// SPDX-License-Identifier: MIT

package serializer_test

import (
	"encoding/json"
	"encoding/xml"
	"testing"

	"github.com/issue9/assert/v2"
	"gopkg.in/yaml.v3"

	"github.com/issue9/web/serializer"
)

var (
	_ json.Marshaler = &serializer.Problem{}
	_ xml.Marshaler  = &serializer.Problem{}
	_ yaml.Marshaler = &serializer.Problem{}
)

func TestProblem_Marshal(t *testing.T) {
	a := assert.New(t, false)

	p1 := serializer.NewProblem()
	p1.Set("type", "400")
	p1.Set("title", "bad request")
	p1.Set("status", 200)

	p2 := serializer.NewProblem()
	p2.Set("type", "400")
	p2.Set("title", "bad request")
	p2.Set("invalid-params", []serializer.InvalidParam{{Name: "n1", Reason: "r1"}, {Name: "n1", Reason: "r2"}})

	t.Run("JSON", func(t *testing.T) {
		data, err := json.Marshal(p1)
		a.NotError(err).
			Equal(string(data), `{"type":"400","title":"bad request","status":200}`)

		data, err = json.Marshal(p2)
		a.NotError(err).
			Equal(string(data), `{"type":"400","title":"bad request","invalid-params":[{"name":"n1","reason":"r1"},{"name":"n1","reason":"r2"}]}`)
	})

	t.Run("XML", func(t *testing.T) {
		data, err := xml.Marshal(p1)
		a.NotError(err).
			Equal(string(data), `<problem xmlns="urn:ietf:rfc:7807"><type>400</type><title>bad request</title><status>200</status></problem>`)

		data, err = xml.Marshal(p2)
		a.NotError(err).
			Equal(string(data), `<problem xmlns="urn:ietf:rfc:7807"><type>400</type><title>bad request</title><invalid-params><i><name>n1</name><reason>r1</reason></i><i><name>n1</name><reason>r2</reason></i></invalid-params></problem>`)
	})

	t.Run("YAML", func(t *testing.T) {
		data, err := yaml.Marshal(p1)
		a.NotError(err).
			Equal(string(data), `type: "400"
title: bad request
status: 200
`)

		data, err = yaml.Marshal(p2)
		a.NotError(err).
			Equal(string(data), `type: "400"
title: bad request
invalid-params:
    - name: n1
      reason: r1
    - name: n1
      reason: r2
`)
	})

}

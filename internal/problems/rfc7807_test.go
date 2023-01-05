// SPDX-License-Identifier: MIT

package problems

import (
	"encoding/json"
	"encoding/xml"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/logs/v4"
)

var (
	_ json.Marshaler = &RFC7807[*ctxDemo]{}
	_ xml.Marshaler  = &RFC7807[*ctxDemo]{}
	_ ctx            = &ctxDemo{}

	testPool = NewRFC7807Pool[*ctxDemo]()
)

type ctxDemo struct{}

func (ctx *ctxDemo) Marshal(status int, body any, problem bool) error {
	return nil
}

func (ctx *ctxDemo) Logs() *logs.Logs {
	return logs.New(logs.NewNopWriter())
}

func TestNewRFC7807(t *testing.T) {
	a := assert.New(t, false)
	p := testPool.New("id", "title", 400)
	a.NotNil(p)
	p.With("instance", "https://example.com/instance/1")

	a.PanicString(func() {
		p.With("instance", "instance")
	}, "存在同名的参数")
	a.PanicString(func() {
		p.With("type", "1111")
	}, "存在同名的参数")

	a.Equal(p.vals[3], "https://example.com/instance/1").
		Equal(p.keys[3], "instance").
		Equal(p.status, 400).
		Equal(p.vals[0], "id").
		Equal(p.vals[1], "title").
		Equal(p.vals[2], 400)
}

func TestRFC7807_Marshal(t *testing.T) {
	a := assert.New(t, false)

	// NOTE: title 因为未调用 apply，所以未翻译，为空。
	p1 := testPool.New("400", "bad request", 200)
	p2 := testPool.New("400", "bad request", 400)
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
		data, err := p1.MarshalForm()
		a.NotError(err).
			Equal(string(data), `status=200&title=bad+request&type=400`)

		data, err = p2.MarshalForm()
		a.NotError(err).
			Equal(string(data), `detail=detail&params%5B0%5D.name=n1&params%5B0%5D.reason=r1&status=400&title=bad+request&type=400`)
	})

	t.Run("HTML", func(t *testing.T) {
		name, v := p1.MarshalHTML()
		a.Equal(name, "problem").
			Equal(v, map[string]any{
				"type":   "400",
				"title":  "bad request",
				"status": 200,
			})

		name, v = p2.MarshalHTML()
		a.Equal(name, "problem").
			Equal(v, map[string]any{
				"type":   "400",
				"title":  "bad request",
				"status": 400,
				"detail": "detail",
				"array":  []string{"a", "bc"},
				"object": &struct{ X string }{X: "x"},
				"params": map[string]string{"n1": "r1"},
			})
	})

	p1.Apply(&ctxDemo{})
	p2.Apply(&ctxDemo{})
}

// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package result

import (
	"encoding/json"
	"encoding/xml"
	"testing"

	"github.com/issue9/assert"
	yaml "gopkg.in/yaml.v2"
)

func TestResultJSONMarshal(t *testing.T) {
	a := assert.New(t)
	a.NotError(NewMessage(400, "400"))

	r := New(400)
	r.Add("field1", "message1")
	r.Add("field2", "message2")

	bs, err := json.Marshal(r)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `{"message":"400","code":400,"detail":[{"field":"field1","message":"message1"},{"field":"field2","message":"message2"}]}`)

	r = New(400)
	bs, err = json.Marshal(r)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `{"message":"400","code":400}`)

	cleanMessage()
}

func TestResultXMLMarshal(t *testing.T) {
	a := assert.New(t)
	a.NotError(NewMessage(400, "400"))

	r := New(400)
	r.Add("field", "message1")
	r.Add("field", "message2")

	bs, err := xml.Marshal(r)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `<result message="400" code="400"><field name="field">message1</field><field name="field">message2</field></result>`)

	r = New(400)
	bs, err = xml.Marshal(r)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `<result message="400" code="400"></result>`)

	cleanMessage()
}

func TestResultYAMLMarshal(t *testing.T) {
	a := assert.New(t)
	a.NotError(NewMessage(400, "400"))

	r := New(400)
	r.Add("field", "message1")
	r.Add("field", "message2")

	bs, err := yaml.Marshal(r)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `message: "400"
code: 400
detail:
- field: field
  message: message1
- field: field
  message: message2
`)

	r = New(400)
	bs, err = yaml.Marshal(r)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `message: "400"
code: 400
`)

	cleanMessage()
}

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

var (
	mimetypeResult = &Result{
		Code:    400,
		Message: "400",
		Detail: []*detail{
			&detail{
				Field:   "field1",
				Message: "message1",
			},
			&detail{
				Field:   "field2",
				Message: "message2",
			},
		},
	}

	simpleMimetypeResult = &Result{
		Code:    400,
		Message: "400",
	}
)

func TestResultJSONMarshal(t *testing.T) {
	a := assert.New(t)

	bs, err := json.Marshal(mimetypeResult)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `{"message":"400","code":400,"detail":[{"field":"field1","message":"message1"},{"field":"field2","message":"message2"}]}`)

	bs, err = json.Marshal(simpleMimetypeResult)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `{"message":"400","code":400}`)

	cleanMessage()
}

func TestResultXMLMarshal(t *testing.T) {
	a := assert.New(t)

	bs, err := xml.Marshal(mimetypeResult)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `<result message="400" code="400"><field name="field1">message1</field><field name="field2">message2</field></result>`)

	bs, err = xml.Marshal(simpleMimetypeResult)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `<result message="400" code="400"></result>`)

	cleanMessage()
}

func TestResultYAMLMarshal(t *testing.T) {
	a := assert.New(t)

	bs, err := yaml.Marshal(mimetypeResult)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `message: "400"
code: 400
detail:
- field: field1
  message: message1
- field: field2
  message: message2
`)

	bs, err = yaml.Marshal(simpleMimetypeResult)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `message: "400"
code: 400
`)

	cleanMessage()
}

func TestResultFormMarshal(t *testing.T) {
	a := assert.New(t)

	bs, err := mimetypeResult.MarshalForm()
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `code=400&detail.field1=message1&detail.field2=message2&message=400`)

	bs, err = simpleMimetypeResult.MarshalForm()
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `code=400&message=400`)

	cleanMessage()
}

// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"testing"

	"github.com/issue9/assert"
)

func TestNewResult(t *testing.T) {
	a := assert.New(t)

	r := NewResult(-2, nil) // 不存在的代码
	a.Equal(r.Code, -1)

	code := http.StatusBadRequest * 1000
	a.NotError(NewMessage(code, "400"))
	r = NewResult(code, nil)
	a.Equal(r.Message, "400").Equal(r.Status(), 400).Equal(r.Code, code)

	clearMesages()
}

func TestResult_Add_HasDetail(t *testing.T) {
	a := assert.New(t)

	code := 400 * 1000
	a.NotError(NewMessage(code, "400"))
	r := NewResult(code, nil)
	a.False(r.HasDetail())

	r.Add("field", "message")
	a.True(r.HasDetail())

	clearMesages()
}

func TestResult_IsError(t *testing.T) {
	a := assert.New(t)

	code := 400 * 1000
	a.NotError(NewMessage(code, "400"))
	r := NewResult(400+500, nil)
	a.True(r.IsError())

	code = 300 * 1000
	a.NotError(NewMessage(code, "400"))
	r = NewResult(code+3, nil)
	a.True(r.IsError())

	// 不存在于 message 中，算是 500 错误
	r = NewResult(200*100+3, nil)
	a.True(r.IsError())

	clearMesages()
}

func TestResultJSONMarshal(t *testing.T) {
	a := assert.New(t)
	a.NotError(NewMessage(400, "400"))

	r := NewResult(400, nil)
	r.Add("field1", "message1")
	r.Add("field2", "message2")

	bs, err := json.Marshal(r)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `{"message":"400","code":400,"detail":[{"field":"field1","message":"message1"},{"field":"field2","message":"message2"}]}`)

	r = NewResult(400, nil)
	bs, err = json.Marshal(r)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `{"message":"400","code":400}`)

	clearMesages()
}

func TestResultXMLMarshal(t *testing.T) {
	a := assert.New(t)
	a.NotError(NewMessage(400, "400"))

	r := NewResult(400, nil)
	r.Add("field", "message1")
	r.Add("field", "message2")

	bs, err := xml.Marshal(r)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `<result message="400" code="400"><field name="field">message1</field><field name="field">message2</field></result>`)

	r = NewResult(400, nil)
	bs, err = xml.Marshal(r)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `<result message="400" code="400"></result>`)

}
